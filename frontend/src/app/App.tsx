import {Alert, Button, Loader, Text} from '@gravity-ui/uikit';
import {Navigate, Route, Routes, useLocation} from 'react-router';
import {useSession} from '../modules/session';
import {HomePage} from '../pages/home/ui/HomePage';
import {LoginPage} from '../pages/login/ui/LoginPage';

function Entry() {
  const session = useSession();
  const location = useLocation();

  if (session.isPending) {
    return (
      <div className="entry-state" role="status">
        <Loader size="m" />
        <Text>Проверяем вход…</Text>
      </div>
    );
  }

  if (session.isError) {
    return (
      <div className="entry-state">
        <div role="alert">
          <Alert
            theme="warning"
            title="Не удалось проверить вход"
            message="Сервис временно недоступен. Попробуйте ещё раз."
          />
        </div>
        <Button view="action" size="l" onClick={() => void session.refetch()}>
          Попробовать снова
        </Button>
      </div>
    );
  }

  if (!session.data) {
    return <LoginPage returnTo={location.pathname} />;
  }

  if (location.pathname === '/login') {
    return <Navigate to="/" replace />;
  }

  return <HomePage profile={session.data} />;
}

export function App() {
  return (
    <div className="app-shell">
      <header className="app-header">
        <a href="/" className="brand" aria-label="Лады — главная">
          <Text variant="header-1">Лады</Text>
          <Text variant="caption-1">ГИТАРНАЯ СТУДИЯ</Text>
        </a>
        <Text variant="caption-2" color="secondary">
          МЕСТО ДЛЯ ВАШЕЙ МУЗЫКИ
        </Text>
      </header>
      <main id="main-content" className="app-main">
        <Routes>
          <Route path="/" element={<Entry />} />
          <Route path="/login" element={<Entry />} />
          <Route path="/songs/:songId" element={<Entry />} />
          <Route path="/songs/:songId/fragments/:fragmentId" element={<Entry />} />
          <Route
            path="*"
            element={
              <section className="entry-state">
                <Text as="h1" variant="header-1">
                  Страница не найдена
                </Text>
                <Button href="/" view="outlined">
                  На главную
                </Button>
              </section>
            }
          />
        </Routes>
      </main>
      <footer className="app-footer">
        <Text variant="caption-2" color="secondary">
          Лады · Личная библиотека Guitar Pro
        </Text>
      </footer>
    </div>
  );
}
