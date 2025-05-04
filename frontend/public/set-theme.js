try {
  console.log('Setting theme');
  const theme = JSON.parse(
    localStorage.getItem('homebox/preferences/location')
  ).theme;
  if (theme) {
    document.documentElement.setAttribute('data-theme', theme);
    document.documentElement.classList.add('theme-' + theme);
  }
} catch (e) {
  console.error('Failed to set theme', e);
}
