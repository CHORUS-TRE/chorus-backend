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