import {lazy, Suspense} from 'react';
import {Navigate, Route, Routes, useLocation} from 'react-router';

import {SessionError, SessionLoading, useSession} from '../../modules/session';
import {AppLayout} from '../ui/AppLayout';

const LoginPage = lazy(() => import('../../modules/session/pages/LoginPage'));
const HomePage = lazy(() => import('../../pages/home/ui/HomePage'));
const NotFoundPage = lazy(
    () => import('../../pages/not-found/ui/NotFoundPage')
);

function Entry() {
    const session = useSession();
    const location = useLocation();

    if (session.isPending) return <SessionLoading />;
    if (session.isError)
        return <SessionError onRetry={() => void session.refetch()} />;
    if (!session.data) return <LoginPage returnTo={location.pathname} />;
    if (location.pathname === '/login') return <Navigate to='/' replace />;

    return <HomePage profile={session.data} />;
}

export function AppRouter() {
    return (
        <AppLayout>
            <Suspense fallback={<SessionLoading />}>
                <Routes>
                    <Route path='/' element={<Entry />} />
                    <Route path='/login' element={<Entry />} />
                    <Route path='/songs/:songId' element={<Entry />} />
                    <Route
                        path='/songs/:songId/fragments/:fragmentId'
                        element={<Entry />}
                    />
                    <Route path='*' element={<NotFoundPage />} />
                </Routes>
            </Suspense>
        </AppLayout>
    );
}
