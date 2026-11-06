import {Text} from '@gravity-ui/uikit';
import {SignInButtons} from '../../../modules/session';
import styles from './LoginPage.module.css';

export function LoginPage({returnTo}: {returnTo: string}) {
  return (
    <div className={styles.layout}>
      <section className={styles.intro} aria-labelledby="intro-title">
        <Text variant="caption-2" className={styles.eyebrow}>
          ЛИЧНАЯ ГИТАРНАЯ СТУДИЯ
        </Text>
        <Text as="h1" variant="display-4" id="intro-title" className={styles.title}>
          Свой ритм.
          <br />
          Свои лады.
        </Text>
        <Text as="p" variant="body-3" className={styles.description}>
          Любимые песни и маленькие шаги к большой музыке.
        </Text>
        <div className={styles.strings} aria-hidden="true">
          <i />
          <i />
          <i />
          <i />
          <i />
          <i />
          <span className={styles.position}>09 — 12</span>
        </div>
        <Text variant="caption-2" className={styles.note}>
          GUITAR PRO · ПЕСНИ · ФРАГМЕНТЫ
        </Text>
      </section>
      <section className={styles.entry} aria-labelledby="login-title">
        <div className={styles.form}>
          <Text variant="caption-2" color="secondary" className={styles.eyebrow}>
            НАЧНЁМ С МУЗЫКИ
          </Text>
          <Text as="h2" variant="header-1" id="login-title">
            Войти в Лады
          </Text>
          <Text as="p" variant="body-2" color="secondary">
            Выберите аккаунт для своей библиотеки Guitar Pro.
          </Text>
          <SignInButtons returnTo={returnTo} />
          <Text as="p" variant="caption-2" color="secondary" className={styles.hint}>
            Каждый аккаунт — отдельная библиотека.
            <br />
            Используйте тот же способ входа, что и раньше.
          </Text>
        </div>
      </section>
    </div>
  );
}
