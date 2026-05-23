/**
 * Converts a boolean to the string "true" or "false".
 * @param {boolean} value
 * @returns {string}
 */
export function toBool(value) {
    return value ? 'true' : 'false';
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
    lines.push(`      - ${key}=${value}`);
}

