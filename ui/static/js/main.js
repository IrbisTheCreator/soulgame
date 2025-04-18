// main.js

// Функция для обработки навигационных ссылок
var navLinks = document.querySelectorAll("nav a");
for (var i = 0; i < navLinks.length; i++) {
    var link = navLinks[i]
    if (link.getAttribute('href') == window.location.pathname) {
        link.classList.add("live");
        break;
    }
}

// Общая функция для обработки событий
function setupEventListeners() {
    // Функция для обработки взятия задания
    document.querySelectorAll('.take-quest-btn').forEach(button => {
        button.addEventListener('click', async () => {
            const questId = button.getAttribute('data-quest-id');
            try {
                const response = await fetch(`/snippet?id=${questId}`, {  // Изменил endpoint на /takequest
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                });
                
                if (response.ok) {
                    // Обновляем интерфейс
                    button.textContent = 'Сдать задание';
                    button.classList.remove('take-quest-btn');
                    button.classList.add('complete-quest-btn');
                    
                    // Удаляем старый обработчик и добавляем новый
                    button.replaceWith(button.cloneNode(true));
                    
                    // Повторно устанавливаем обработчики для всех кнопок
                    setupEventListeners();
                } else {
                    const errorData = await response.json();
                    alert(errorData.message || 'Ошибка при взятии задания');
                }
            } catch (error) {
                console.error('Ошибка:', error);
                alert('Произошла ошибка при отправке запроса');
            }
        });
    });

    // Функция для обработки сдачи задания
    document.querySelectorAll('.complete-quest-btn').forEach(button => {
        button.addEventListener('click', async () => {
            const questId = button.getAttribute('data-quest-id');
            try {
                const response = await fetch(`/completequest?id=${questId}`, {
                    method: 'POST',
                    headers: {
                        'Content-Type': 'application/json',
                    },
                });
                
                if (response.ok) {
                    // Обновляем интерфейс
                    button.textContent = 'Ожидайте проверки';
                    button.disabled = true;
                } else {
                    const errorData = await response.json();
                    alert(errorData.message || 'Ошибка при сдаче задания');
                }
            } catch (error) {
                console.error('Ошибка:', error);
                alert('Произошла ошибка при отправке запроса');
            }
        });
    });
}

document.querySelectorAll('.checkin').forEach(button => {
    button.addEventListener('click', async () => {
        const questId = parseInt(button.getAttribute('data-quest-id'));
        
        button.disabled = true;
        const originalText = button.textContent;
        button.textContent = 'Загрузка...';
        
        try {
            const response = await fetch('/checkin', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({ quest_id: questId })
            });
            
            if (!response.ok) {
                const errorData = await response.json();
                throw new Error(errorData?.error || 'Ошибка сервера');
            }

            let data;
            try {
                data = await response.json();
            } catch (e) {
                throw new Error('Неверный формат ответа от сервера');
            }
            
            if (!data || !data.success) {
                throw new Error(data?.message || 'Неизвестная ошибка');
            }
            
            button.textContent = 'Выполнено';
            button.disabled = true;
            showNotification(data.message);

            if (data.souls !== undefined) {
                updateSoulsCount(data.souls);
            }
            
        } catch (error) {
            console.error('Ошибка:', error);
            button.disabled = false;
            button.textContent = originalText;
            alert("Ты уже выполнил задание сегодня");
        }
    });
});

function updateSoulsCount(newCount) {
    const soulsElements = document.querySelectorAll('nav span strong');
    soulsElements.forEach(element => {
        element.textContent = `У вас ${newCount} душ`;
    });
}

document.querySelectorAll('.buy-btn').forEach(button => {
    button.addEventListener('click', async function() {
        const itemId = this.getAttribute('data-item-id');
        const clubSelect = document.querySelector(`.club-select[data-item-id="${itemId}"]`);
        const clubId = clubSelect.value;
        
        if (!clubId) {
            alert('Пожалуйста, выберите клуб');
            return;
        }
        
        this.disabled = true;
        this.textContent = 'Покупка...';
        
        try {
            const response = await fetch('/purchase', {
                method: 'POST',
                headers: {
                    'Content-Type': 'application/json',
                },
                body: JSON.stringify({
                    item_id: parseInt(itemId),
                    club_id: parseInt(clubId)
                })
            });

            // Сначала проверяем статус ответа
            if (!response.ok) {
                // Пытаемся прочитать как JSON, если не получается - как текст
                let errorData;
                try {
                    errorData = await response.json();
                } catch {
                    errorData = { error: await response.text() };
                }
                throw new Error(errorData.error || `HTTP error! status: ${response.status}`);
            }

            const result = await response.json();
            alert('Покупка успешна!');
            location.reload();
            
        } catch (error) {
            console.error('Детали ошибки:', error);
            alert(`Недостаточно душ`);
        } finally {
            button.disabled = false;
            button.textContent = 'Купить';
        }
    });
});

function showNotification(message) {
    alert(message); // или кастомная реализация уведомления
}


// Инициализация обработчиков при загрузке страницы
document.addEventListener('DOMContentLoaded', setupEventListeners);