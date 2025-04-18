document.addEventListener("DOMContentLoaded", function () {
    const phoneInput = document.getElementById("phone");

    if (phoneInput) {
        phoneInput.addEventListener("input", function (e) {
            let value = e.target.value.replace(/\D/g, ""); // Удаляем все нецифровые символы
            if (value.length > 11) value = value.slice(0, 11); // Ограничиваем длину 11 цифрами

            // Форматируем номер телефона
            let formattedValue = "+7 ";
            if (value.length > 1) {
                formattedValue += "(" + value.slice(1, 4);
            }
            if (value.length > 4) {
                formattedValue += ") " + value.slice(4, 7);
            }
            if (value.length > 7) {
                formattedValue += "-" + value.slice(7, 9);
            }
            if (value.length > 9) {
                formattedValue += "-" + value.slice(9, 11);
            }

            e.target.value = formattedValue; // Устанавливаем отформатированное значение
        });

        // Очистка поля, если маска не заполнена полностью
        phoneInput.addEventListener("blur", function (e) {
            if (e.target.value.replace(/\D/g, "").length < 11) {
                e.target.value = ""; // Очищаем поле, если номер неполный
            }
        });
    }
});