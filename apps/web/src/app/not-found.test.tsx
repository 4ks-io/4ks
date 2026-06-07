import React from 'react';
import { renderToStaticMarkup } from 'react-dom/server';
import NotFound from './not-found';

describe('NotFound', () => {
  it('renders a Forks branded 404 page with the logo and recovery links', () => {
    const html = renderToStaticMarkup(<NotFound />);

    expect(html).toContain('src="/logo.svg"');
    expect(html).toContain('alt="4ks.io"');
    expect(html).toContain('Forks');
    expect(html).toContain('404');
    expect(html).toContain('Page not found');
    expect(html).toContain('href="/"');
    expect(html).toContain('href="/explore"');
    expect(html).toContain('MuiButton-containedSecondary');
  });
});
