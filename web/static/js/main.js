document.addEventListener('DOMContentLoaded', () => {
    const form = document.getElementById('shortenForm');
    
    if(form) {
        form.addEventListener('submit', async (e) => {
            e.preventDefault();
            
            const urlInput = document.getElementById('urlInput');
            const customCode = document.getElementById('customCode');
            const resultDiv = document.getElementById('result');
            const submitBtn = form.querySelector('button');

            submitBtn.disabled = true;
            submitBtn.textContent = 'Сокращаем...';

            try {
                const response = await fetch('/api/links', {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                        'Authorization': `Bearer ${localStorage.getItem('token') || ''}`
                    },
                    body: JSON.stringify({
                        original_url: urlInput.value,
                        custom_code: customCode.value.trim() || undefined
                    })
                });

                const data = await response.json();
                
                if (!response.ok) {
                    throw new Error(data.error || 'Ошибка сервера');
                }

                resultDiv.innerHTML = `
                    <div class="success">
                        ✅ Короткая ссылка: 
                        <a href="/${data.short_code}" target="_blank" class="short-link">
                            ${window.location.host}/${data.short_code}
                        </a>
                    </div>
                `;

                urlInput.value = '';
                customCode.value = '';

            } catch (err) {
                resultDiv.innerHTML = `
                    <div class="error">
                        ❌ Ошибка: ${err.message}
                    </div>
                `;
            } finally {
                submitBtn.disabled = false;
                submitBtn.textContent = 'Сократить';
            }
        });
    }
});