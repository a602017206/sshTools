import { describe, it } from 'node:test';
import assert from 'node:assert/strict';
import { nextIndex, prevIndex, wrapIndex } from '../assets/js/carousel.js';

describe('carousel index helpers', () => {
  it('wraps forward and backward', () => {
    assert.equal(nextIndex(3, 4), 0);
    assert.equal(prevIndex(0, 4), 3);
    assert.equal(wrapIndex(-1, 4), 3);
  });
});
