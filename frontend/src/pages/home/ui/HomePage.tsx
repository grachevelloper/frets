import {Text} from '@gravity-ui/uikit';
import {SignOutButton} from '../../../modules/session';
import type {Profile} from '../../../modules/session';

export function HomePage({profile}: {profile: Profile}) {
  return (
    <section className="entry-state">
      <Text variant="caption-2" color="secondary">
        ЛИЧНАЯ БИБЛИОТЕКА
      </Text>
      <Text as="h1" variant="header-1">
        Вы вошли, {profile.nickname}
      </Text>
      <Text as="p" variant="body-2" color="secondary">
        Добро пожаловать в вашу гитарную студию.
      </Text>
      <SignOutButton />
    </section>
  );
}
