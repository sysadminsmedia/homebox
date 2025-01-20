try {
    const theme = JSON.parse(localStorage.getItem("homebox/preferences/location")).theme
    if (theme) document.documentElement.setAttribute("data-theme", theme)
} catch(e) {/* */}