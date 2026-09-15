import { CONTAINER_PORT } from './utils.js';

// Every Homebox image declares a HEALTHCHECK, so sidecars can wait for the app
// to actually answer requests rather than merely for its container to start.
const dependsOnHomebox = [
    '    depends_on:',
    '      homebox:',
    '        condition: service_healthy',
];

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
            ...dependsOnHomebox,
            '    ports:',
            '      - "80:80"',
            '      - "443:443"',
            '      - "443:443/udp"',
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
            ...dependsOnHomebox,
            '    ports:',
            '      - "80:80"',
            '      - "443:443"',
            '    volumes:',
            '      - ./nginx.conf:/etc/nginx/conf.d/default.conf:ro',
        ].join('\n');
    }

    if (state.proxyType === 'cloudflare') {
        return [
            '  cloudflared:',
            '    image: cloudflare/cloudflared:latest',
            '    restart: unless-stopped',
            ...dependsOnHomebox,
            '    command: tunnel --no-autoupdate run',
            '    environment:',
            // Read from the .env file / shell rather than baked into the compose
            // file, so the tunnel credential never lands in version control.
            '      - TUNNEL_TOKEN=${CLOUDFLARE_TUNNEL_TOKEN:?set CLOUDFLARE_TUNNEL_TOKEN in a .env file}',
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
    if (state.proxyType !== 'traefik') {
        return '';
    }

    const host = state.hostname || 'homebox.example.com';

    return [
        '    labels:',
        '      - "traefik.enable=true"',
        // Without an explicit port Traefik has to guess which exposed port to
        // route to, which breaks as soon as the container exposes more than one.
        `      - "traefik.http.services.homebox.loadbalancer.server.port=${CONTAINER_PORT}"`,
        '      - "traefik.http.services.homebox.loadbalancer.passhostheader=true"',
        `      - "traefik.http.routers.homebox-http.rule=Host(\`${host}\`)"`,
        '      - "traefik.http.routers.homebox-http.entrypoints=web"',
        '      - "traefik.http.routers.homebox-http.middlewares=homebox-https-redirect"',
        `      - "traefik.http.routers.homebox.rule=Host(\`${host}\`)"`,
        '      - "traefik.http.routers.homebox.entrypoints=websecure"',
        '      - "traefik.http.routers.homebox.tls=true"',
        '      - "traefik.http.routers.homebox.middlewares=homebox-sslheader"',
        '      - "traefik.http.middlewares.homebox-https-redirect.redirectscheme.scheme=https"',
        '      - "traefik.http.middlewares.homebox-sslheader.headers.customrequestheaders.X-Forwarded-Proto=https"',
    ].join('\n');
}
