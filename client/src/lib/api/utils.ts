export function setTokenInLocalStorage(token: string, hours: number) {
    const expiryTime = new Date().getTime() + hours * 60 * 60 * 1000;
    localStorage.setItem("roomToken", JSON.stringify({ value: token, expiry: expiryTime }));
}

export function getTokenFromLocalStorage(): string | null {
    const item = localStorage.getItem("roomToken");
    if (!item) return null;

    const { value, expiry } = JSON.parse(item);
    if (new Date().getTime() > expiry) {
        localStorage.removeItem("roomToken");
        return null;
    }
    return value;
}