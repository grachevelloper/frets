import {Provider} from '@gravity-ui/uikit';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {render, screen} from '@testing-library/react';
import userEvent from '@testing-library/user-event';
import {MemoryRouter} from 'react-router';
import {describe, expect, it, vi} from 'vitest';
import {App} from './App';

const profile = {id: 'user-1', nickname: 'Николай', email: null, createdAt: '2026-11-06'};

function renderApp(path = '/') {
  const client = new QueryClient({defaultOptions: {queries: {retry: false}}});

  return render(
    <Provider theme="light" lang="ru">
      <QueryClientProvider client={client}>
        <MemoryRouter initialEntries={[path]}>
          <App />
        </MemoryRouter>
      </QueryClientProvider>
    </Provider>,
  );
}

describe('session entry', () => {
  it('shows loading until the session is known', () => {
    vi.stubGlobal(
      'fetch',
      vi.fn(() => new Promise(() => {})),
    );
    renderApp();

    expect(screen.getByRole('status')).toHaveTextContent('Проверяем вход');
    expect(screen.queryByRole('link', {name: 'Войти с Яндекс ID'})).not.toBeInTheDocument();
  });

  it('preserves the fragment path in both OAuth links', async () => {
    vi.stubGlobal('fetch', vi.fn().mockResolvedValue(new Response('{}', {status: 401})));
    renderApp('/songs/song-1/fragments/fragment-1');

    const yandex = await screen.findByRole('link', {name: 'Войти с Яндекс ID'});
    expect(yandex).toHaveAttribute(
      'href',
      '/api/v1/auth/login/yandex?next=%2Fsongs%2Fsong-1%2Ffragments%2Ffragment-1',
    );
    expect(screen.getByRole('link', {name: 'Войти с Google'})).toHaveAttribute(
      'href',
      '/api/v1/auth/login/google?next=%2Fsongs%2Fsong-1%2Ffragments%2Ffragment-1',
    );
  });

  it('retries an unavailable service instead of treating it as signed out', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(new Response('{"code":"unavailable"}', {status: 503}))
      .mockResolvedValueOnce(Response.json(profile));
    vi.stubGlobal('fetch', fetch);
    renderApp();

    expect(await screen.findByRole('alert')).toHaveTextContent('Не удалось проверить вход');
    expect(screen.queryByRole('link', {name: 'Войти с Google'})).not.toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', {name: 'Попробовать снова'}));
    expect(await screen.findByRole('heading', {name: 'Вы вошли, Николай'})).toBeInTheDocument();
  });

  it('keeps the profile on logout failure and clears it after a successful retry', async () => {
    const fetch = vi
      .fn()
      .mockResolvedValueOnce(Response.json(profile))
      .mockResolvedValueOnce(new Response('{}', {status: 503}))
      .mockResolvedValueOnce(new Response(null, {status: 204}));
    vi.stubGlobal('fetch', fetch);
    renderApp();

    await userEvent.click(await screen.findByRole('button', {name: 'Выйти'}));
    expect(await screen.findByRole('alert')).toHaveTextContent('Не удалось выйти');
    expect(screen.getByRole('heading', {name: 'Вы вошли, Николай'})).toBeInTheDocument();
    await userEvent.click(screen.getByRole('button', {name: 'Выйти'}));
    expect(await screen.findByRole('link', {name: 'Войти с Google'})).toBeInTheDocument();
    expect(fetch).toHaveBeenLastCalledWith(
      '/api/v1/auth/logout',
      expect.objectContaining({method: 'POST', credentials: 'same-origin'}),
    );
  });
});
