// Backend configuration
const BACKEND_PORT = '3008';

/**
 * Returns the appropriate backend URL based on how the frontend is being accessed.
 * If accessed via localhost, uses localhost.
 * If accessed via IP or domain, uses that same address.
 */
export const getBackendUrl = (): string => {
  // Get the current hostname and protocol
  const protocol = window.location.protocol;
  const hostname = window.location.hostname;

  // If we're accessing from localhost, use localhost
  if (hostname === 'localhost' || hostname === '127.0.0.1') {
    return `http://localhost:${BACKEND_PORT}`;
  }

  // Otherwise, use the current hostname
  return `${protocol}//${hostname}:${BACKEND_PORT}`;
};

export const API_BASE_URL = getBackendUrl();

