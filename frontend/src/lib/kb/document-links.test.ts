import { describe, expect, it } from 'vitest';
import { resolveDocumentLink } from './document-links';

describe('Knowledge document links', () => {
  const current = 'packages/review/start.md';
  it.each([
    ['scripts.md', 'packages/review/scripts.md'],
    ['./scripts.md', 'packages/review/scripts.md'],
    ['../other/scripts.md', 'packages/other/scripts.md'],
    ['/pages/start.md', 'pages/start.md'],
    ['my%20script.md', 'packages/review/my script.md']
  ])('resolves %s inside the workspace', (href, path) => {
    expect(resolveDocumentLink(href, current)).toEqual({ path, fragment: '' });
  });
  it('retains fragments without passing query parameters to the file API', () => {
    expect(resolveDocumentLink('scripts.md?workspace=other#first-part', current))
      .toEqual({ path: 'packages/review/scripts.md', fragment: 'first-part' });
    expect(resolveDocumentLink('#first-part', current)).toEqual({ path: current, fragment: 'first-part' });
  });
  it.each(['https://example.com/file.md', '//example.com/file.md', 'mailto:hello@example.com', '/tasks', 'video.mp4'])('leaves ordinary URLs unchanged: %s', href => {
    expect(resolveDocumentLink(href, current)).toBeNull();
  });
  it.each(['../../../private.md', '%2e%2e/%2e%2e/%2e%2e/private.md', 'bad%ZZ.md', '..\\private.md'])('blocks malformed or escaping file links: %s', href => {
    expect(resolveDocumentLink(href, current)).toBe('blocked');
  });
});
