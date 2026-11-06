import {describe, expect, it} from 'vitest';

import {getReturnPath} from './returnPath';

const UNSAFE_PATHS = [
    'https://example.com',
    '//example.com',
    '/login',
    '/songs/../login',
    '/songs/one\\two',
];

describe('getReturnPath', () => {
    it.each(UNSAFE_PATHS)('rejects %s', (path) => {
        const expected = '/';

        const result = getReturnPath(path);

        expect(result).toBe(expected);
    });

    it('keeps song and fragment routes', () => {
        const path = '/songs/song-1/fragments/fragment-1';

        const result = getReturnPath(path);

        expect(result).toBe(path);
    });
});
