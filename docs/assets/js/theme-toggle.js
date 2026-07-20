(() => {
  const storageKey = "goenums-theme";
  const root = document.documentElement;
  const button = document.querySelector(".theme-toggle");

  const setTheme = (theme) => {
    root.dataset.theme = theme;
    localStorage.setItem(storageKey, theme);
  };

  const preferredTheme = localStorage.getItem(storageKey) ||
    (window.matchMedia("(prefers-color-scheme: dark)").matches ? "dark" : "light");
  setTheme(preferredTheme);

  button?.addEventListener("click", () => {
    setTheme(root.dataset.theme === "dark" ? "light" : "dark");
  });
})();
