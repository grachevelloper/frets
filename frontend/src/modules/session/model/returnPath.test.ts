import {describe, expect, it} from 'vitest';
import {getReturnPath} from './returnPath';

const unsafePaths = [
  'https://example.com',
  '//example.com',
  '/login',
  '/songs/../login',
  '/songs/one\\two',
];

describe('getReturnPath', () => {
  it.each(unsafePaths)('rejects %s', (path) => {
    expect(getReturnPath(path)).toBe('/');
  });

  it('keeps song and fragment routes', () => {
    expect(getReturnPath('/songs/song-1/fragments/fragment-1')).toBe(
      '/songs/song-1/fragments/fragment-1',
    );
  });
});
