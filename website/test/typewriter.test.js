import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { buildTypeSequence } from '../assets/js/typewriter.js';

describe('buildTypeSequence', () => {
  it('types and clears each line', () => {
    const frames = buildTypeSequence(['ab'], { pause: 1, clearPause: 1 });
    assert.ok(frames.includes('a'));
    assert.ok(frames.includes('ab'));
    assert.equal(frames[frames.length - 1], '');
  });
});
