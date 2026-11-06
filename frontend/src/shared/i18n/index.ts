import {I18N} from '@gravity-ui/i18n';

import ru from './ru.json';

const instance = new I18N();
instance.registerKeysets('ru', {app: ru});
instance.setLang('ru');

export function i18n(key: keyof typeof ru, params?: Record<string, string>) {
    return instance.i18n('app', key, params);
}
