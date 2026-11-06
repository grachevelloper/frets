import {Button} from '@gravity-ui/uikit';
import {getLoginHref} from '../api/session';
import styles from './Session.module.css';

export function SignInButtons({returnTo}: {returnTo: string}) {
  return (
    <div className={styles.buttons}>
      <Button href={getLoginHref('yandex', returnTo)} view="action" size="xl" width="max">
        Войти с Яндекс ID
      </Button>
      <Button href={getLoginHref('google', returnTo)} view="outlined" size="xl" width="max">
        Войти с Google
      </Button>
    </div>
  );
}
