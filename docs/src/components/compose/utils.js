/**
 * Converts a boolean to the string "true" or "false".
 * @param {boolean} value
 * @returns {string}
 */
export function toBool(value) {
    return value ? 'true' : 'false';
}

// Values the designer interpolates are free text typed by the reader - passwords,
// paths, connection strings, OIDC secrets. Pasted raw they can break the document
// (a colon in a password) or, worse, change it silently (everything after a "#"
// is read as a comment). Every state-derived value goes through yamlScalar.
const PLAIN_SAFE = /^[A-Za-z0-9_][A-Za-z0-9_./@+=-]*$/;
const NEEDS_DOUBLE_QUOTES = /[\p{Cc}\p{Cf}]/u;

/**
 * Encodes an arbitrary string as a YAML scalar, leaving unambiguous values in
 * plain style so the generated file stays readable.
 *
 * @param {string|number|boolean} value
 * @returns {string}
 */
export function yamlScalar(value) {
    const raw = String(value);

    if (PLAIN_SAFE.test(raw)) {
        return raw;
    }
    // Control characters have no single-quoted representation; JSON escaping is
    // valid YAML double-quoted style.
    if (NEEDS_DOUBLE_QUOTES.test(raw)) {
        return JSON.stringify(raw);
    }
    // Single-quoted style needs no escaping beyond doubling the quote itself.
    return `'${raw.replace(/'/g, "''")}'`;
}

/**
 * Builds a `KEY=value` entry for a service's `environment:` sequence, encoded so
 * the value cannot alter the surrounding YAML.
 *
 * @param {string} key
 * @param {string|number|boolean} value
 * @returns {string}
 */
export function envLine(key, value) {
    return `      - ${yamlScalar(`${key}=${value}`)}`;
}

/**
 * Appends an env-var line to `lines` only when `value` is non-empty.
 * @param {string[]} lines
 * @param {string} key
 * @param {string} value
 */
export function pushEnv(lines, key, value) {
    if (!value) {
        return;
    }
    lines.push(envLine(key, value));
}

/** Port the Homebox container listens on (EXPOSE in every image variant). */
export const CONTAINER_PORT = '7745';
