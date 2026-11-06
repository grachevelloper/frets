import {Alert, Button} from '@gravity-ui/uikit';
import {useLogout} from '../model/useSession';
import styles from './Session.module.css';

export function SignOutButton() {
  const logout = useLogout();

  return (
    <div className={styles.buttons}>
      {logout.isError && (
        <div role="alert">
          <Alert theme="danger" title="Не удалось выйти" message="Попробуйте ещё раз." />
        </div>
      )}
      <Button view="outlined" size="l" loading={logout.isPending} onClick={() => logout.mutate()}>
        Выйти
      </Button>
    </div>
  );
}
