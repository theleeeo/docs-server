window.onload = function () {
    setupVersionSelect();
};

function setupVersionSelect() {
    fetch('/docs/versions')
        .then(response => response.json())
        .then(versions => {
            var dropdown = document.getElementById('dropdown');
            dropdown.onchange = handleVersionSelect;

            versions.forEach(version => {
                var option = document.createElement('option');
                option.text = version;
                option.value = version;

                dropdown.add(option);
            });
        })
        .catch(error => console.error('Error:', error));
    }

function handleVersionSelect() {
    const version = document.getElementById('dropdown').value;
    fetch(`/docs/version/${version}/roles`)
        .then(response => response.json())
        .then(roles => {
            buttonContainer = document.getElementById('button-container');
            buttonContainer.innerHTML = '';
            roles.forEach(role => {
                button = document.createElement('button');
                button.innerHTML = role;

                button.onclick = function () {
                    window.location.href = `/docs/${version}/${role}`;
                };

                buttonContainer.appendChild(button);
            });
        })
        .catch(error => console.error('Error:', error));
}
