function login(event) {
    event.preventDefault();

    hideError();

    const username = document.getElementById('username').value;
    const password = document.getElementById('password').value;

    const data = {
        username: username,
        password: password
    };

    fetch('/api/rest/v1/authentication/login', {
        method: 'POST',
        headers: {
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    })
        .then(response => response.json())
        .then(data => {
            if (data?.result?.token) {
                const urlParams = new URLSearchParams(window.location.search);
                const callbackUrl = urlParams.get('callback_url') || '/';
                window.location.href = callbackUrl;
            } else if (data?.message) {
                displayError(data.message);
            }
        })
        .catch(error =>{ 
            console.error('Error:', error);
            displayError('An unexpected error occurred. Please try again later.' + error);
        });
}

function hideError() {
    const errorElement = document.getElementById('error');
    errorElement.textContent = '';
    errorElement.parentElement.classList.add('hidden');
}

function displayError(message) {
    const errorElement = document.getElementById('error');
    errorElement.textContent = message;
    errorElement.parentElement.classList.remove('hidden');
}