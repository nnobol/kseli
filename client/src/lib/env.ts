const env = import.meta.env;

export const API_KEY = env.VITE_API_KEY;
export const useMocks = env.VITE_ENV === 'local';