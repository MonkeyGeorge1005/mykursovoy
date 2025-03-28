function handleSubmit(e) {
    e.preventDefault();
    const email = document.getElementById("email").value;
    const password = document.getElementById("password").value;
    const username = document.getElementById("username").value;

    if (!email || !password || !username) {
        alert("Пожалуйста, заполните все поля.");
        return;
    }

    const userData = {
        email: email,
        password: password,
        username: username,
    };

    fetch("/register", {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(userData),
    })
    .then(response => {
        if (!response.ok) {
            throw new Error("Network response was not ok");
        }
        return response.json();
    })
    .then(data => {
        if (data.message === "User registered successfully") {
            window.location.href = "/login";
        } else {
            alert("Ошибка регистрации: " + data.error);
        }
    })
    .catch(error => {
        console.error("Ошибка при отправке данных:", error);
        alert("Произошла ошибка. Попробуйте позже.");
    });
}