export type DocumentLink = { path: string; fragment: string } | 'blocked' | null;

// Resolve file links against the document, never the application's web route.
export function resolveDocumentLink(href: string, currentPath: string): DocumentLink {
  const value = href.trim();
  if (!value || /^[a-z][a-z0-9+.-]*:/i.test(value) || value.startsWith('//')) return null;
  const hash = value.indexOf('#');
  const target = (hash < 0 ? value : value.slice(0, hash)).split('?')[0];
  let fragment: string;
  let decoded: string;
  try {
    fragment = decodeURIComponent(hash < 0 ? '' : value.slice(hash + 1));
    decoded = decodeURIComponent(target);
  } catch { return 'blocked'; }
  if (!target && hash >= 0) return { path: currentPath, fragment };
  if (!/\.md$/i.test(decoded)) return null;
  if (/[\\\x00-\x1f]/.test(decoded)) return 'blocked';
  const parts = decoded.startsWith('/') ? [] : currentPath.split('/').slice(0, -1);
  for (const part of decoded.split('/')) {
    if (!part || part === '.') continue;
    if (part === '..') {
      if (!parts.length) return 'blocked';
      parts.pop();
    } else parts.push(part);
  }
  return { path: parts.join('/'), fragment };
}
