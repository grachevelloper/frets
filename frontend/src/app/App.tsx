import {Provider} from '@gravity-ui/uikit';
import {QueryClient, QueryClientProvider} from '@tanstack/react-query';
import {useState} from 'react';
import {BrowserRouter} from 'react-router';

import {AppRouter} from './router/AppRouter';

export function App() {
    const [queryClient] = useState(
        () =>
            new QueryClient({
                defaultOptions: {
                    queries: {retry: false},
                    mutations: {retry: false},
                },
            })
    );

    return (
        <Provider theme='light' lang='ru'>
            <QueryClientProvider client={queryClient}>
                <BrowserRouter>
                    <AppRouter />
                </BrowserRouter>
            </QueryClientProvider>
        </Provider>
    );
}
