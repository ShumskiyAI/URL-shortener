document.getElementById('shorten-form').addEventListener('submit', async (e) => {
    e.preventDefault();
    const input = document.getElementById('url-input');
    const result = document.querySelector('.result');
    try {
        const res = await fetch('/shorten', {
            method: 'POST',
            headers: {'Content-Type': 'application/json'},
            body: JSON.stringify({url: input.value}),
        });
        if (!res.ok) {
            const text = await res.text();
            result.innerHTML = `Ошибка: <span>${text}</span>`;
            return;
        }
        const data = await res.json();
        result.innerHTML = `Готово: <a href="/${data.short}" target="_blank">/${data.short}</a>`;
    } catch (err) {
        result.innerHTML = `Ошибка: <span>${err.message}</span>`;
    }
});