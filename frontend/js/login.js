document.getElementById("loginForm").addEventListener("submit", async (e) => {
    e.preventDefault();
    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;

    try {
        const response = await fetch("/login", {
            method: "POST",
            headers: { "Content-Type": "application/json", "Accept": "application/json" },
            body: JSON.stringify({ email, password }),
            credentials: 'include'
        });

        const data = await response.json();
        if (data.success) {
            window.location.href = "/employee";
        } else {
            showError("Неверные данные");
        }
    } catch (error) {
        showError("Ошибка сети");
    }
});

function showError(message) {
    const errorElement = document.getElementById("error-message");
    if (errorElement) {
        errorElement.textContent = message;
        errorElement.style.display = "block";
    } else {
        alert(message);
    }
}