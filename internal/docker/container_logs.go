package docker

import (
	"context"
	"encoding/binary"
	"fmt"
	"io"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

func (c *Client) GetContainerLogs(ctx context.Context, id string, tail int) (string, error) {
	reqURL := containerLogsURL(c.baseURL, id, tail, false)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return "", err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return "", err
	}
	defer resp.Body.Close()
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		return "", fmt.Errorf("docker: get container logs failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	decoded, err := readDockerMultiplexed(resp.Body)
	if err != nil {
		return "", err
	}
	return string(decoded), nil
}

func (c *Client) FollowContainerLogs(ctx context.Context, id string, tail int) (io.ReadCloser, error) {
	reqURL := containerLogsURL(c.baseURL, id, tail, true)
	req, err := http.NewRequestWithContext(ctx, http.MethodGet, reqURL, nil)
	if err != nil {
		return nil, err
	}
	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, err
	}
	if resp.StatusCode != http.StatusOK {
		body, _ := io.ReadAll(io.LimitReader(resp.Body, 512))
		_ = resp.Body.Close()
		return nil, fmt.Errorf("docker: follow container logs failed: %s: %s", resp.Status, strings.TrimSpace(string(body)))
	}
	rc := &demuxLogReadCloser{
		ctx:  ctx,
		body: resp.Body,
		mr:   newMultiplexReader(resp.Body),
	}
	go func() {
		<-ctx.Done()
		_ = rc.Close()
	}()
	return rc, nil
}

func containerLogsURL(baseURL, id string, tail int, follow bool) string {
	v := url.Values{}
	v.Set("stdout", "1")
	v.Set("stderr", "1")
	v.Set("timestamps", "0")
	if tail > 0 {
		v.Set("tail", strconv.Itoa(tail))
	}
	if follow {
		v.Set("follow", "1")
	}
	return fmt.Sprintf("%s/containers/%s/logs?%s", baseURL, id, v.Encode())
}

func readDockerMultiplexed(r io.Reader) ([]byte, error) {
	mr := newMultiplexReader(r)
	return io.ReadAll(mr)
}

type multiplexReader struct {
	r           io.Reader
	pending     []byte
	payloadLeft int
	hdr         [8]byte
	hdrFilled   int
}

func newMultiplexReader(r io.Reader) *multiplexReader {
	return &multiplexReader{r: r}
}

func (m *multiplexReader) Read(p []byte) (int, error) {
	for {
		if len(m.pending) > 0 {
			n := copy(p, m.pending)
			m.pending = m.pending[n:]
			return n, nil
		}
		if m.payloadLeft > 0 {
			toRead := m.payloadLeft
			if toRead > len(p) {
				toRead = len(p)
			}
			n, err := io.ReadFull(m.r, p[:toRead])
			m.payloadLeft -= n
			if err != nil {
				return n, err
			}
			return n, nil
		}
		if err := m.readFrameHeader(); err != nil {
			return 0, err
		}
	}
}

func (m *multiplexReader) readFrameHeader() error {
	for m.hdrFilled < 8 {
		n, err := m.r.Read(m.hdr[m.hdrFilled:])
		m.hdrFilled += n
		if err != nil {
			return err
		}
	}
	size := binary.BigEndian.Uint32(m.hdr[4:8])
	m.hdrFilled = 0
	m.payloadLeft = int(size)
	return nil
}

type demuxLogReadCloser struct {
	ctx      context.Context
	body     io.ReadCloser
	mr       *multiplexReader
	closeOnce bool
}

func (d *demuxLogReadCloser) Read(p []byte) (int, error) {
	return d.mr.Read(p)
}

func (d *demuxLogReadCloser) Close() error {
	if d.closeOnce {
		return nil
	}
	d.closeOnce = true
	return d.body.Close()
}
