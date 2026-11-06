import {Provider} from '@gravity-ui/uikit';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import type {PropsWithChildren} from 'react';
import {BrowserRouter} from 'react-router';

const queryClient = new QueryClient({
  defaultOptions: {queries: {retry: false}, mutations: {retry: false}},
});

export function AppProviders({children}: PropsWithChildren) {
  return (
    <Provider theme="light" lang="ru">
      <QueryClientProvider client={queryClient}>
        <BrowserRouter>{children}</BrowserRouter>
      </QueryClientProvider>
    </Provider>
  );
}
