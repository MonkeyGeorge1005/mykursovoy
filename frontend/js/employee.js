//загрузка из куки токна, аватарку и имя пользователя, а также вставка плейлистов из бд
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

    const userAvatar = document.getElementById('userAvatar');
    const usernameElement = document.getElementById('username');
    usernameElement.addEventListener('mouseenter', async () => {
        try {
            if (!colorCache[userAvatar.src]) {
                colorCache[userAvatar.src] = await getDominantColor(userAvatar.src);
            }
            usernameElement.style.color = colorCache[userAvatar.src];
        } catch (error) {
            console.error('Error processing color:', error);
            usernameElement.style.color = '#1db954';
        }
    });

    usernameElement.addEventListener('mouseleave', () => {
        usernameElement.style.color = '';
    });

    const notReadList = document.getElementById('notReadList');
    const readList = document.getElementById('readList');
    const notRead = document.querySelector('.not-read');
    const Read = document.querySelector('.read');
    let countread = 0;
    let countnotread = 0;

    try {
        const messageResponse = await fetch("/api/messages");
        if (!messageResponse.ok) {
            Read.classList.add('hidden');
            notRead.classList.add('hidden');
        }
        else if (messageResponse.ok){
            const messages = await messageResponse.json();

            if (messages.length === 0) {
                alert('Нет доступных сообщений');
                return;
            }

            for (const message of messages) {
                const messageShow = document.createElement('div');
                messageShow.classList.add(message.is_read ? 'read-message' : 'not-read-message');
                messageShow.dataset.messageId = message.bid_id;

                messageShow.innerHTML = `
                    <div class="message-title">
                        <span class="message-title-fio">${message.employee_name}</span>
                    </div>
                    <div class="message-info">
                        <span class="message-age">Возраст: ${message.age}</span>
                        <span class="message-job">Должность: ${message.job_title}</span>
                        <span class="message-subdivision">Подразделение: ${message.subdivision}</span>
                        <span class="message-languages">Языки: ${message.languages.map(lang => `${lang.language} (${lang.proficiency})`).join(', ')}</span>
                        <span class="message-educations">Образование: ${message.educations.map(educ => `${educ.name} (${educ.place})`).join(', ')}</span>
                        <span class="experience">Общий стаж: ${message.overall_experience} лет</span>
                        <span class="s-p-experience">Научно-технический стаж: ${message.s_p_experience} лет</span>
                        <div class="confirm-deny">
                            <button type="button" class="confirm-button" data-message-id="${message.bid_id}">Принять</button>
                            <button type="button" class="deny-button" data-message-id="${message.bid_id}">Отклонить</button>
                        </div>
                    </div>
                `;

                if (message.is_read) {
                    countread++;
                    readList.appendChild(messageShow);
                } else {
                    countnotread++;
                    notReadList.appendChild(messageShow);
                }
            }
        }

        if (countread === 0) {
            Read.classList.add('hidden');
        } else {
            Read.classList.remove('hidden');
        }

        if(countnotread === 0) {
            notRead.classList.add('hidden');
        } else {
            notRead.classList.remove('hidden');
        }

        const messageForRead = await fetch("/api/messagesread")
        if (!messageResponse.ok) throw new Error('Ошибка обновления статуса сообщений');

        document.body.addEventListener('click', async (event) => {
            const target = event.target;
            // Обработка кнопки "Принять"
            if (target.classList.contains('confirm-button')) {
                const messageId = target.dataset.messageId;
                try {
                    const response = await fetch(`/api/accept-application/${messageId}`, {
                        method: 'POST',
                        headers: { 'Content-Type': 'application/json' },
                        credentials: 'include'
                    });
                    if (!response.ok) {
                        const errorData = await response.json();
                        throw new Error(errorData.error || 'Ошибка принятия заявки');
                    }
                    location.reload();
                } catch (error) {
                    console.error('Ошибка:', error);
                    alert('Не удалось принять заявку');
                }
            }
    
            // Обработка кнопки "Отклонить"
            if (target.classList.contains('deny-button')) {
                const messageId = target.dataset.messageId;
                try {
                    const response = await fetch(`/api/reject-application/${messageId}`, {
                        method: 'DELETE',
                        headers: { 'Content-Type': 'application/json' },
                        credentials: 'include'
                    });
                    if (!response.ok) {
                        const errorData = await response.json();
                        throw new Error(errorData.error || 'Ошибка отклонения заявки');
                    }
                    location.reload();
                } catch (error) {
                    console.error('Ошибка:', error);
                    alert('Не удалось отклонить заявку');
                }
            }
        });

    } catch (error) {
        console.error('Ошибка загрузки данных сообщений:', error);
    }
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

document.querySelectorAll('.menu-item').forEach(btn => {
    btn.addEventListener('click', () => {
        document.querySelectorAll('.menu-item').forEach(item => {
            item.classList.remove('active');
        });
        
        btn.classList.add('active');
    });
});

document.getElementById('mail').addEventListener('click', () => {
    const sidebar = document.querySelector('.table-employee-sidebar');
    const jobTitleSidebar = document.querySelector('.table-job-title-sidebar');
    const subdivisionSidebar = document.querySelector('.table-subdivision-sidebar');

    sidebar.classList.remove('active');
    jobTitleSidebar.classList.remove('active');
    subdivisionSidebar.classList.remove('active');
});

document.getElementById('employee_table').addEventListener('click', () => {
    loadTableData("employees");
    const sidebar = document.querySelector('.table-employee-sidebar');
    const jobTitleSidebar = document.querySelector('.table-job-title-sidebar');
    const subdivisionSidebar = document.querySelector('.table-subdivision-sidebar');

    sidebar.classList.add('active');
    jobTitleSidebar.classList.remove('active');
    subdivisionSidebar.classList.remove('active');

    const table = document.getElementById('employeesTable');
    const headers = table.querySelectorAll('th');
    let currentSortColumn = null;
    let isAscending = false;

    headers.forEach((header, index) => {
        header.addEventListener('click', () => {
            const rows = Array.from(table.querySelectorAll('tbody tr'));
            const dataType = header.getAttribute('data-type');

            if (currentSortColumn === index) {
                isAscending = !isAscending;
            } else {
                isAscending = true;
                currentSortColumn = index;
            }

            rows.sort((a, b) => {
                const aValue = a.cells[index].textContent;
                const bValue = b.cells[index].textContent;

                if (dataType === 'number') {
                    return isAscending ? aValue - bValue : bValue - aValue;
                } else {
                    return isAscending ? aValue.localeCompare(bValue) : bValue.localeCompare(aValue);
                }
            });

            const tbody = table.querySelector('tbody');
            tbody.append(...rows);

            headers.forEach(h => h.classList.remove('sorted-asc', 'sorted-desc'));
            header.classList.add(isAscending ? 'sorted-asc' : 'sorted-desc');
        });
    });

    document.getElementById('applyEmployeeFilter').addEventListener('click', () => {
        const id = document.getElementById('idFilter').value;
        const fio = document.getElementById('fioFilter').value;
        const age = document.getElementById('ageFilter').value;
        const jobTitle = document.getElementById('jobTitleFilter').value;
        const subdivision = document.getElementById('subdivisionFilter').value;
        const overall = document.getElementById('overallFilter').value;
        const s_p = document.getElementById('s_pFilter').value;

        const filters = {
            id: id,
            fio: fio,
            age: age,
            job_title_id: jobTitle,
            subdivision_id: subdivision,
            overall_experience: overall,
            s_p_experience: s_p,
        };

        loadTableData("employees", filters);
    });
});

document.getElementById('job_title_table').addEventListener('click', () => {
    loadJobTitleTableData("job_title");
    const sidebar = document.querySelector('.table-employee-sidebar');
    const jobTitleSidebar = document.querySelector('.table-job-title-sidebar');
    const subdivisionSidebar = document.querySelector('.table-subdivision-sidebar');

    sidebar.classList.remove('active');
    jobTitleSidebar.classList.add('active');
    subdivisionSidebar.classList.remove('active');

    const table = document.getElementById('jobTitleTable');
    const headers = table.querySelectorAll('th');
    let currentSortColumn = null;
    let isAscending = false;

    headers.forEach((header, index) => {
        header.addEventListener('click', () => {
            const rows = Array.from(table.querySelectorAll('tbody tr'));
            const dataType = header.getAttribute('data-type');

            if (currentSortColumn === index) {
                isAscending = !isAscending;
            } else {
                isAscending = true;
                currentSortColumn = index;
            }

            rows.sort((a, b) => {
                const aValue = a.cells[index].textContent;
                const bValue = b.cells[index].textContent;

                if (dataType === 'number') {
                    return isAscending ? aValue - bValue : bValue - aValue;
                } else {
                    return isAscending ? aValue.localeCompare(bValue) : bValue.localeCompare(aValue);
                }
            });

            const tbody = table.querySelector('tbody');
            tbody.append(...rows);

            headers.forEach(h => h.classList.remove('sorted-asc', 'sorted-desc'));
            header.classList.add(isAscending ? 'sorted-asc' : 'sorted-desc');
        });
    });

    document.getElementById('applyJobTitleFilter').addEventListener('click', () => {
        const id = document.getElementById('idJobTitleFilter').value;
        const name = document.getElementById('nameJobTitleFilter').value;
        const count = document.getElementById('countJobTitleFilter').value;

        const filters = {
            id: id,
            name: name,
            count: count,
        };

        loadJobTitleTableData("job_title", filters);
    });
});

document.getElementById('subdiv_table').addEventListener('click', () => {
    loadSubdivisionTableData("subdivision");
    const sidebar = document.querySelector('.table-employee-sidebar');
    const jobTitleSidebar = document.querySelector('.table-job-title-sidebar');
    const subdivisionSidebar = document.querySelector('.table-subdivision-sidebar');

    sidebar.classList.remove('active');
    jobTitleSidebar.classList.remove('active');
    subdivisionSidebar.classList.add('active');

    const table = document.getElementById('subdivisionTable');
    const headers = table.querySelectorAll('th');
    let currentSortColumn = null;
    let isAscending = false;

    headers.forEach((header, index) => {
        header.addEventListener('click', () => {
            const rows = Array.from(table.querySelectorAll('tbody tr'));
            const dataType = header.getAttribute('data-type');

            if (currentSortColumn === index) {
                isAscending = !isAscending;
            } else {
                isAscending = true;
                currentSortColumn = index;
            }

            rows.sort((a, b) => {
                const aValue = a.cells[index].textContent;
                const bValue = b.cells[index].textContent;

                if (dataType === 'number') {
                    return isAscending ? aValue - bValue : bValue - aValue;
                } else {
                    return isAscending ? aValue.localeCompare(bValue) : bValue.localeCompare(aValue);
                }
            });

            const tbody = table.querySelector('tbody');
            tbody.append(...rows);

            headers.forEach(h => h.classList.remove('sorted-asc', 'sorted-desc'));
            header.classList.add(isAscending ? 'sorted-asc' : 'sorted-desc');
        });
    });

    document.getElementById('applySubdivisionFilter').addEventListener('click', () => {
        const id = document.getElementById('idSubdivisionFilter').value;
        const name = document.getElementById('nameSubdivisionFilter').value;

        const filters = {
            id: id,
            name: name,
        };

        loadSubdivisionTableData("subdivision", filters);
    });
});

function getDominantColor(imageUrl) {
    return new Promise((resolve, reject) => {
        const canvas = document.createElement('canvas');
        const ctx = canvas.getContext('2d');
        const img = new Image();
        img.crossOrigin = 'Anonymous';
        img.src = imageUrl;
        img.onload = () => {
            canvas.width = img.width / 4;
            canvas.height = img.height / 4;
            ctx.drawImage(img, 0, 0, canvas.width, canvas.height);
            const imageData = ctx.getImageData(0, 0, canvas.width, canvas.height);
            const data = imageData.data;
            const colorCounts = {};
            for (let i = 0; i < data.length; i += 4) {
                const r = data[i];
                const g = data[i + 1];
                const b = data[i + 2];
                if (r + g + b > 170 && r + g + b < 680) {
                    const colorKey = `${r}-${g}-${b}`;
                    if (colorCounts[colorKey]) {
                        colorCounts[colorKey]++;
                    } else {
                        colorCounts[colorKey] = 1;
                    }
                }
            }
            let dominantColor = [0, 0, 0];
            let maxCount = 0;
            for (const key in colorCounts) {
                const count = colorCounts[key];
                if (count > maxCount) {
                    maxCount = count;
                    dominantColor = key.split('-').map(Number);
                }
            }
            resolve(`rgb(${dominantColor[0]}, ${dominantColor[1]}, ${dominantColor[2]})`);
        };
        img.onerror = () => reject('Ошибка загрузки изображения');
    });
}

const colorCache = {};

async function loadTableData(tableType, filters = {}) {
    try {
        const queryParams = new URLSearchParams();
        for (const [key, value] of Object.entries(filters)) {
            if (value) {
                queryParams.append(key, value);
            }
        }

        const response = await fetch(`/api/${tableType}/get?${queryParams.toString()}`);
        if (!response.ok) {
            throw new Error('Ошибка загрузки данных');
        }
        const data = await response.json();

        const tbody = document.querySelector(`#${tableType}Table tbody`);
        tbody.innerHTML = '';

        data.forEach(item => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${item.id}</td>
                <td>${item.fio}</td>
                <td>${item.age}</td>
                <td>${item.job_title_id}</td>
                <td>${item.subdivision_id}</td>
                <td>${item.overall_experience}</td>
                <td>${item.s_p_experience}</td>
                <td>
                    <button class="changebutton" onclick="editRow(${item.id})">Изменить</button>
                    <button class="deletebutton" onclick="deleteRow(${item.id})">Удалить</button>
                </td>
            `;
            tbody.appendChild(row);
        });
    } catch (error) {
        console.error('Ошибка загрузки данных:', error);
    }
}

async function loadJobTitleTableData(tableType, filters = {}) {
    try {
        const queryParams = new URLSearchParams();
        for (const [key, value] of Object.entries(filters)) {
            if (value) {
                queryParams.append(key, value);
            }
        }

        const response = await fetch(`/api/${tableType}/get?${queryParams.toString()}`);
        if (!response.ok) {
            throw new Error('Ошибка загрузки данных');
        }
        const data = await response.json();

        const tbody = document.querySelector(`#jobTitleTable tbody`);
        tbody.innerHTML = '';

        data.forEach(item => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${item.id}</td>
                <td>${item.name}</td>
                <td>${item.count}</td>
            `;
            tbody.appendChild(row);
        });
    } catch (error) {
        console.error('Ошибка загрузки данных:', error);
    }
}

async function loadSubdivisionTableData(tableType, filters = {}) {
    try {
        const queryParams = new URLSearchParams();
        for (const [key, value] of Object.entries(filters)) {
            if (value) {
                queryParams.append(key, value);
            }
        }

        const response = await fetch(`/api/${tableType}/get?${queryParams.toString()}`);
        if (!response.ok) {
            throw new Error('Ошибка загрузки данных');
        }
        const data = await response.json();

        const tbody = document.querySelector(`#${tableType}Table tbody`);
        tbody.innerHTML = '';

        data.forEach(item => {
            const row = document.createElement('tr');
            row.innerHTML = `
                <td>${item.id}</td>
                <td>${item.name}</td>
            `;
            tbody.appendChild(row);
        });
    } catch (error) {
        console.error('Ошибка загрузки данных:', error);
    }
}

let currentEditedId = null;

async function editRow(itemId) {
    try {
        const response = await fetch(`/api/employees/${itemId}`);
        if (!response.ok) throw new Error('Ошибка загрузки данных');
        
        const item = await response.json();
        console.log('Received item:', item);

        document.getElementById('editFio').value = item.fio || '';
        document.getElementById('editAge').value = item.age || '';
        document.getElementById('editJobTitle').value = item.job_title_id || '';
        document.getElementById('editSubdivision').value = item.subdivision_id || '';
        document.getElementById('editOverallExp').value = item.overall_experience || '';
        document.getElementById('editSPExp').value = item.s_p_experience || '';
        
        currentEditedId = itemId;
        document.getElementById('editFormEmployees').style.display = 'block';
    } catch (error) {
        console.error('Ошибка редактирования:', error);
        alert('Не удалось загрузить данные для редактирования');
    }
}

async function saveEdit() {
    try {
        const loadingIndicator = document.createElement('div');
        loadingIndicator.className = 'loading';
        loadingIndicator.textContent = 'Сохранение...';
        document.getElementById('editFormEmployees').appendChild(loadingIndicator);

        const editedItem = {
            fio: document.getElementById('editFio').value.trim(),
            age: parseInt(document.getElementById('editAge').value, 10) || 0,
            job_title_id: parseInt(document.getElementById('editJobTitle').value, 10) || 0,
            subdivision_id: parseInt(document.getElementById('editSubdivision').value, 10) || 0,
            overall_experience: parseInt(document.getElementById('editOverallExp').value, 10) || 0,
            s_p_experience: parseInt(document.getElementById('editSPExp').value, 10) || 0
        };

        if (!editedItem.fio || editedItem.age < 0) {
            throw new Error('Проверьте введенные данные');
        }

        const response = await fetch(`/api/employees/${currentEditedId}`, {
            method: 'PUT',
            headers: { 'Content-Type': 'application/json' },
            body: JSON.stringify(editedItem)
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Ошибка обновления');
        }

        const id = document.getElementById('idFilter').value;
        const fio = document.getElementById('fioFilter').value;
        const age = document.getElementById('ageFilter').value;
        const jobTitle = document.getElementById('jobTitleFilter').value;
        const subdivision = document.getElementById('subdivisionFilter').value;
        const overall = document.getElementById('overallFilter').value;
        const s_p = document.getElementById('s_pFilter').value;

        
        const filters = {
            id: id,
            fio: fio,
            age: age,
            job_title_id: jobTitle,
            subdivision_id: subdivision,
            overall_experience: overall,
            s_p_experience: s_p,
        };

        document.getElementById('editFormEmployees').style.display = 'none';
        loadTableData("employees", filters);
    } catch (error) {
        console.error('Ошибка:', error);
        alert(`Ошибка: ${error.message}`);
    } finally {
        const loadingIndicator = document.querySelector('.loading');
        if (loadingIndicator) {
            loadingIndicator.remove();
        }
    }
}

async function deleteRow(itemId) {
    if (!confirm('Вы уверены, что хотите удалить этого сотрудника?')) return;

    try {
        const response = await fetch(`/api/employees/${itemId}`, {
            method: 'DELETE',
            headers: { 'Content-Type': 'application/json' },
        });

        if (!response.ok) {
            const errorData = await response.json();
            throw new Error(errorData.error || 'Ошибка удаления');
        }

        const id = document.getElementById('idFilter').value;
        const fio = document.getElementById('fioFilter').value;
        const age = document.getElementById('ageFilter').value;
        const jobTitle = document.getElementById('jobTitleFilter').value;
        const subdivision = document.getElementById('subdivisionFilter').value;
        const overall = document.getElementById('overallFilter').value;
        const s_p = document.getElementById('s_pFilter').value;
        
        const filters = {
            id: id,
            fio: fio,
            age: age,
            job_title_id: jobTitle,
            subdivision_id: subdivision,
            overall_experience: overall,
            s_p_experience: s_p,
        };

        loadTableData("employees", filters);
    } catch (error) {
        console.error('Ошибка:', error);
        alert(`Ошибка: ${error.message}`);
    }
}

document.getElementById('editFormEmployees').querySelector('button').addEventListener('click', (e) => {
    e.preventDefault();
    saveEdit();
});

window.addEventListener('click', (e) => {
    const form = document.getElementById('editFormEmployees');
    if (e.target === form) form.style.display = 'none';
});

document.getElementById('exportDataButton').addEventListener('click', function() {
    // Получаем данные из таблицы
    const table = document.getElementById("employeesTable");

    // Заголовки таблицы
    const headers = Array.from(table.querySelectorAll("thead th")).map(th => th.innerText.trim());

    // Строки данных
    const rows = Array.from(table.querySelectorAll("tbody tr"));
    const jsonData = [];

    rows.forEach(row => {
        const rowData = Array.from(row.querySelectorAll("td")).map(cell => cell.innerText.trim());
        const rowObj = {};

        headers.forEach((header, index) => {
            rowObj[header] = rowData[index];
        });

        jsonData.push(rowObj);
    });

    // Преобразуем данные в строку JSON
    const jsonString = JSON.stringify(jsonData, null, 2);

    // Создаем Blob для JSON файла
    const blob = new Blob([jsonString], { type: 'application/json' });
    const url = URL.createObjectURL(blob);

    // Создаем ссылку для скачивания файла
    const link = document.createElement("a");
    link.setAttribute("href", url);
    link.setAttribute("download", "employees_data.json"); // Имя файла
    document.body.appendChild(link); // Добавляем ссылку в DOM
    link.click(); // Программно кликаем по ссылке для скачивания
    link.remove(); // Удаляем ссылку после скачивания
});