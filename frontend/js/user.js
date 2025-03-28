document.addEventListener('DOMContentLoaded', async () => {
    try {
        const roleResponse = await fetch("/employee", {
            method: "GET",
            headers: {
                "Content-Type": "application/json",
            },
            credentials: 'include'
        });

        if (!roleResponse.ok) {
            const errorData = await roleResponse.json();
            if (errorData.error === "No role") {
                window.location.href = "/login";
            } else {
                throw new Error(errorData.error || "Ошибка сервера");
            }
        }


        const userResponse = await fetch('/api/user', {
            credentials: 'include'
        });
        const userData = await userResponse.json();
        const userAvatar = document.getElementById('userAvatar');
        const usernameElement = document.getElementById('username');

        userAvatar.src = userData.logoURL;
        usernameElement.textContent = userData.user;

        userAvatar.onerror = () => {
            userAvatar.src = 'https://i.imgur.com/k8NBJSm.jpg';
        };

    } catch (error) {
        console.error('Ошибка загрузки данных пользователя:', error);
        alert('Не удалось загрузить данные пользователя');
    }

    try {
        const jobTitleResponse = await fetch('/api/job-titles');
        if (!jobTitleResponse.ok) throw new Error('Ошибка загрузки должностей');
        const jobTitles = await jobTitleResponse.json();

        const jobTitleSelect = document.getElementById('job_title');
        jobTitles.forEach(title => {
            const option = document.createElement('option');
            option.value = title.id;
            option.textContent = title.name;
            jobTitleSelect.appendChild(option);
        });

        const subdivisionResponse = await fetch('/api/subdivisions');
        if (!subdivisionResponse.ok) throw new Error('Ошибка загрузки подразделений');
        const subdivisions = await subdivisionResponse.json();

        const subdivisionSelect = document.getElementById('subdivision');
        subdivisions.forEach(subdivision => {
            const option = document.createElement('option');
            option.value = subdivision.id;
            option.textContent = subdivision.name;
            subdivisionSelect.appendChild(option);
        });
    } catch (error) {
        console.error('Ошибка при загрузке данных:', error);
        alert('Не удалось загрузить данные для выпадающих списков');
    }

    const languageResponse = await fetch('/api/languages');
    if (!languageResponse.ok) throw new Error('Ошибка загрузки языков');
    const languages = await languageResponse.json();

    // Добавление языков
    const languagesContainer = document.getElementById('languages-container');
    const addLanguageButton = document.getElementById('addLanguage');
    let languageCounter = 0;

    addLanguageButton.addEventListener('click', () => {
        const languageDiv = document.createElement('div');
        languageDiv.classList.add('dynamic-input-group');

        // Выбор языка
        const languageSelect = document.createElement('select');
        languageSelect.classList.add('input');
        languageSelect.name = `language_${languageCounter}`;
        languageSelect.innerHTML = `<option value="" disabled selected>Выберите язык</option>`;
        languages.forEach(lang => {
            const option = document.createElement('option');
            option.value = lang.id;
            option.textContent = lang.language;
            languageSelect.appendChild(option);
        });

        // Выбор уровня владения
        const proficiencySelect = document.createElement('select');
        proficiencySelect.classList.add('input');
        proficiencySelect.classList.add('language');
        proficiencySelect.language = `proficiency_${languageCounter}`;
        proficiencySelect.innerHTML = `
            <option value="" disabled selected>Выберите уровень</option>
            <option value="A1">A1</option>
            <option value="A2">A2</option>
            <option value="B1">B1</option>
            <option value="B2">B2</option>
            <option value="C1">C1</option>
            <option value="C2">C2</option>
        `;
        proficiencySelect.name = `proficiency_${languageCounter}`;
        

        const removeButton = document.createElement('button');
        removeButton.type = 'button';
        removeButton.textContent = 'Удалить';
        removeButton.classList.add('remove-button');
        removeButton.addEventListener('click', () => {
            languagesContainer.removeChild(languageDiv);
        });

        languageDiv.appendChild(languageSelect);
        languageDiv.appendChild(proficiencySelect);
        languageDiv.appendChild(removeButton);

        languagesContainer.appendChild(languageDiv);
        languageCounter++;
    });

    const educationResponse = await fetch('/api/educations');
    if (!educationResponse.ok) throw new Error('Ошибка загрузки образований');
    const educations = await educationResponse.json();

    // Добавление образования
    const educationsContainer = document.getElementById('educations-container');
    const addEducationButton = document.getElementById('addEducation');
    let educationCounter = 0;

    addEducationButton.addEventListener('click', () => {
        const educationDiv = document.createElement('div');
        educationDiv.classList.add('dynamic-input-group');

        // Выбор учебного заведения
        const educationSelect = document.createElement('select');
        educationSelect.classList.add('input');
        educationSelect.name = `education_${educationCounter}`;
        educationSelect.innerHTML = `<option value="" disabled selected>Выберите учебное заведение</option>`;
        educations.forEach(edu => {
            const option = document.createElement('option');
            option.value = edu.id;
            option.textContent = edu.name;
            educationSelect.appendChild(option);
        });

        // Выбор места получения образования
        const educationPlaceInput = document.createElement('input');
        educationPlaceInput.classList.add('input');
        educationPlaceInput.classList.add('language');
        educationPlaceInput.type = 'text';
        educationPlaceInput.placeholder = 'Место получения образования';
        educationPlaceInput.name = `education_place_${educationCounter}`;

        // Кнопка удаления
        const removeButton = document.createElement('button');
        removeButton.type = 'button';
        removeButton.textContent = 'Удалить';
        removeButton.classList.add('remove-button');
        removeButton.addEventListener('click', () => {
            educationsContainer.removeChild(educationDiv);
        });

        educationDiv.appendChild(educationSelect);
        educationDiv.appendChild(educationPlaceInput);
        educationDiv.appendChild(removeButton);

        educationsContainer.appendChild(educationDiv);
        educationCounter++;
    });
});

document.querySelector('.logout').addEventListener('click', async (event) => {
    event.preventDefault();
    try {
        const response = await fetch('/logout', {
            method: 'POST',
            credentials: 'include'
        });

        if (!response.ok) {
            throw new Error('Ошибка при выходе');
        }
        window.location.href = '/login';
    } catch (error) {
        console.error('Ошибка при выходе:', error);
        alert('Не удалось выйти');
    }
});

document.getElementById('applicationForm').addEventListener('submit', async (e) => {
    e.preventDefault();

    const over_exp = document.getElementById('overall_experience').value;
    const s_p_exp = document.getElementById('s_p_experience').value;
    const age = document.getElementById('age').value;
    if (over_exp < s_p_exp){
        document.getElementById('overall_experience').value = '';
        document.getElementById('s_p_experience').value = '';
        return;
    }

    const job = parseInt(document.getElementById('job_title').value);
    if (job === 2){
        if(s_p_exp < 3){
            document.getElementById('s_p_experience').value = '';
            return;
        }
    }

    if(age <= 0) {
        document.getElementById('age').value = '';
        return;
    }



    const formData = {
        fio: document.getElementById('fio').value,
        age: parseInt(document.getElementById('age').value),
        overall_experience: parseInt(document.getElementById('overall_experience').value),
        s_p_experience: parseInt(document.getElementById('s_p_experience').value),
        job_title_id: parseInt(document.getElementById('job_title').value),
        subdivision_id: parseInt(document.getElementById('subdivision').value),
        languages: [],
        educations: []
    };

    const languageContainers = document.querySelectorAll('#languages-container .dynamic-input-group');
    languageContainers.forEach(container => {
        const languageID = parseInt(container.querySelector('select[name^="language_"]').value);
        const proficiency = container.querySelector('select[name^="proficiency_"]').value;
        formData.languages.push({ language_id: languageID, proficiency });
    });

    const educationContainers = document.querySelectorAll('#educations-container .dynamic-input-group');
    educationContainers.forEach(container => {
        const educationID = parseInt(container.querySelector('select[name^="education_"]').value);
        const place = container.querySelector('input[name^="education_place_"]').value;
        formData.educations.push({ education_id: educationID, place });
    });

    try {
        const response = await fetch('/api/submit-application', {
            method: 'POST',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(formData)
        });

        if (!response.ok) {
            throw new Error('Ошибка отправки заявки');
        }

        alert('Заявка успешно отправлена!');
        document.getElementById('applicationForm').reset();
    } catch (error) {
        console.error('Ошибка:', error);
        alert('Не удалось отправить заявку');
    }
});