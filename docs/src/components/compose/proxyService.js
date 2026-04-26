/**
 * Returns a YAML service block string for the selected proxy type, or an
 * empty string when no proxy is configured.
 *
 * @param {{ proxyType: string }} state
 * @returns {string}
 */
export function proxyServiceBlock(state) {
    if (state.proxyType === 'caddy') {
        return [
            '  caddy:',
            '    image: caddy:2-alpine',
            '    restart: unless-stopped',
            '    depends_on:',
            '      - homebox',
            '    ports:',
            '      - 80:80',
            '      - 443:443',
            '    volumes:',
            '      - ./Caddyfile:/etc/caddy/Caddyfile:ro',
            '      - caddy-data:/data',
            '      - caddy-config:/config',
        ].join('\n');
    }

    if (state.proxyType === 'nginx') {
        return [
            '  nginx:',
            '    image: nginx:alpine',
            '    restart: unless-stopped',
            '    depends_on:',
            '      - homebox',
            '    ports:',
            '      - 80:80',
            '      - 443:443',
            '    volumes:',
            '      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro',
        ].join('\n');
    }

    if (state.proxyType === 'cloudflare') {
        return [
            '  cloudflared:',
            '    image: cloudflare/cloudflared:latest',
            '    restart: unless-stopped',
            '    depends_on:',
            '      - homebox',
            '    command: tunnel --no-autoupdate run --token ${CLOUDFLARE_TUNNEL_TOKEN}',
        ].join('\n');
    }

    return '';
}

/**
 * Returns a YAML labels block string for the selected proxy type, or an
 * empty string when no proxy is configured.
 *
 * @param {{ proxyType: string, hostname: string }} state
 * @returns {string}
 */
export function getLabels(state) {
    if (state.proxyType === 'traefik') {
        return [
            '    labels:',
            '      traefik.enable: true',
            '      traefik.http.routers.homebox-http.rule: Host(`' + state.hostname + '`)',
            '      traefik.http.routers.homebox-http.entrypoints: web',
            '      traefik.http.routers.homebox.rule: Host(`' + state.hostname + '`)',
            '      traefik.http.routers.homebox.entrypoints: websecure',
            '      traefik.http.routers.homebox.tls: true',
            '      traefik.http.services.homebox.loadbalancer.passhostheader: true',
            '      traefik.http.middlewares.treaefik-https-redirect.redirectscheme.scheme: https',
            '      traefik.http.middlewares.sslheader.headers.customrequestheaders.X-Forwarded-Proto: https',
        ].join('\n');
    }
}