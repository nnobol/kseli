export function setItemInLocalStorage(key: string, item: string, hours: number) {
    const expiryTime = new Date().getTime() + hours * 60 * 60 * 1000;
    localStorage.setItem(key, JSON.stringify({ value: item, expiry: expiryTime }));
}

export function setItemInSessionStorage(key: string, item: string, hours: number) {
    const expiryTime = new Date().getTime() + hours * 60 * 60 * 1000;
    sessionStorage.setItem(key, JSON.stringify({ value: item, expiry: expiryTime }));
}


export function getItemFromLocalStorage(key: string): string | null {
    const item = localStorage.getItem(key);
    if (!item) return null;

    let parsed;
    try {
        parsed = JSON.parse(item);
    } catch (e) {
        localStorage.removeItem(key);
        return null;
    }

    const { value, expiry } = parsed;

    if (!value || !expiry || new Date().getTime() > expiry) {
        localStorage.removeItem(key);
        return null;
    }

    return value;
}

export function getItemFromSessionStorage(key: string): string | null {
    const item = sessionStorage.getItem(key);
    if (!item) return null;

    let parsed;
    try {
        parsed = JSON.parse(item);
    } catch (e) {
        sessionStorage.removeItem(key);
        return null;
    }

    const { value, expiry } = parsed;

    if (!value || !expiry || new Date().getTime() > expiry) {
        sessionStorage.removeItem(key);
        return null;
    }

    return value;
}