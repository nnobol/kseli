import type { PageLoad } from './$types';

export const load: PageLoad = async () => {
    sessionStorage.clear();

    return {};
};