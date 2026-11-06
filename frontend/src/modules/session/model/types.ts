export type Profile = {
  id: string;
  nickname: string;
  email: string | null;
  createdAt: string;
};

export type LoginProvider = 'yandex' | 'google';
