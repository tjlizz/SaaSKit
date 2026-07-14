const STORAGE_KEY = 'saaskit-current-application-id';

export function getCurrentApplicationId() {
  return localStorage.getItem(STORAGE_KEY) ?? '';
}

export function setCurrentApplicationId(id: string) {
  if (id) {
    localStorage.setItem(STORAGE_KEY, id);
  } else {
    localStorage.removeItem(STORAGE_KEY);
  }
  window.dispatchEvent(
    new CustomEvent('saaskit:application-changed', { detail: id }),
  );
}
