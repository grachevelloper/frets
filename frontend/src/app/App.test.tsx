import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {describe, expect, it, vi} from 'vitest';

import {App} from './App';

const PROFILE = {
    id: 'user-1',
    nickname: 'Николай',
    email: 'nick@example.test',
    createdAt: '2026-11-06',
};

function renderApp(path = '/') {
    window.history.replaceState(null, '', path);
    return render(<App />);
}

describe('session entry', () => {
    it('shows loading until the session is known', () => {
        vi.stubGlobal(
            'fetch',
            vi.fn(() => new Promise(() => {}))
        );

        renderApp();

        expect(screen.getByRole('status')).toHaveTextContent('Проверяем вход');
        expect(
            screen.queryByRole('link', {name: 'Войти с Яндекс ID'})
        ).not.toBeInTheDocument();
    });

    it('preserves the fragment path in both OAuth links', async () => {
        vi.stubGlobal(
            'fetch',
            vi.fn().mockResolvedValue(new Response('{}', {status: 401}))
        );

        renderApp('/songs/song-1/fragments/fragment-1');
        const yandex = await screen.findByRole('link', {
            name: 'Войти с Яндекс ID',
        });

        expect(yandex).toHaveAttribute(
            'href',
            '/api/v1/auth/login/yandex?next=%2Fsongs%2Fsong-1%2Ffragments%2Ffragment-1'
        );
        expect(
            screen.getByRole('link', {name: 'Войти с Google'})
        ).toHaveAttribute(
            'href',
            '/api/v1/auth/login/google?next=%2Fsongs%2Fsong-1%2Ffragments%2Ffragment-1'
        );
    });

    it('shows a service failure without treating the visitor as signed out', async () => {
        vi.stubGlobal(
            'fetch',
            vi.fn().mockResolvedValue(new Response('{}', {status: 503}))
        );

        renderApp();
        const alert = await screen.findByRole('alert');

        expect(alert).toHaveTextContent('Не удалось проверить вход');
        expect(
            screen.queryByRole('link', {name: 'Войти с Google'})
        ).not.toBeInTheDocument();
    });

    it('recovers the profile after retrying a service failure', async () => {
        vi.stubGlobal(
            'fetch',
            vi
                .fn()
                .mockResolvedValueOnce(new Response('{}', {status: 503}))
                .mockResolvedValueOnce(Response.json(PROFILE))
        );
        renderApp();
        const retry = await screen.findByRole('button', {
            name: 'Попробовать снова',
        });

        await userEvent.click(retry);

        expect(
            await screen.findByRole('heading', {name: 'Вы вошли, Николай'})
        ).toBeInTheDocument();
    });

    it('keeps the profile on logout failure', async () => {
        vi.stubGlobal(
            'fetch',
            vi
                .fn()
                .mockResolvedValueOnce(Response.json(PROFILE))
                .mockResolvedValueOnce(new Response('{}', {status: 503}))
        );
        renderApp();
        const logout = await screen.findByRole('button', {name: 'Выйти'});

        await userEvent.click(logout);

        expect(await screen.findByRole('alert')).toHaveTextContent(
            'Не удалось выйти'
        );
        expect(
            screen.getByRole('heading', {name: 'Вы вошли, Николай'})
        ).toBeInTheDocument();
    });

    it('clears the profile on successful logout', async () => {
        const fetch = vi
            .fn()
            .mockResolvedValueOnce(Response.json(PROFILE))
            .mockResolvedValueOnce(new Response(null, {status: 204}));
        vi.stubGlobal('fetch', fetch);
        renderApp();
        const logout = await screen.findByRole('button', {name: 'Выйти'});

        await userEvent.click(logout);

        expect(
            await screen.findByRole('link', {name: 'Войти с Google'})
        ).toBeInTheDocument();
        expect(fetch).toHaveBeenLastCalledWith(
            '/api/v1/auth/logout',
            expect.objectContaining({
                method: 'POST',
                credentials: 'same-origin',
            })
        );
    });
});
