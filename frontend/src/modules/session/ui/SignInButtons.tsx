import {Flex} from '@gravity-ui/uikit';

import {i18n} from '../../../shared/i18n';
import googleButton from '../assets/google-sign-in.svg';
import yandexButton from '../assets/yandex-sign-in.svg';
import {getLoginHref} from '../lib/getLoginHref';

import styles from './SignInButtons.module.css';

export function SignInButtons({returnTo}: {returnTo: string}) {
    return (
        <Flex
            direction='column'
            gap='spacing-3'
            alignItems='center'
            role='group'
            aria-label={i18n('sign-in-with')}
        >
            <a
                href={getLoginHref('yandex', returnTo)}
                aria-label={i18n('yandex')}
                className={styles.link}
            >
                <img
                    src={yandexButton}
                    alt=''
                    width='260'
                    height='56'
                    className={styles.artwork}
                />
            </a>
            <a
                href={getLoginHref('google', returnTo)}
                aria-label={i18n('google')}
                className={styles.link}
            >
                <img
                    src={googleButton}
                    alt=''
                    width='180'
                    height='40'
                    className={styles.artwork}
                />
            </a>
        </Flex>
    );
}
