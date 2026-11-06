export function getReturnPath(path: string) {
  if (!/^\/songs\/[^/?#\\]+(?:\/fragments\/[^/?#\\]+)?$/.test(path)) {
    return '/';
  }

  const segments = path.split('/');

  return segments.some((segment) => segment === '.' || segment === '..') ? '/' : path;
}
