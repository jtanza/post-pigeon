// cant issue a DELETE on a form action, unfortunately. this is our work around
function sendDeletePost() {
    const formData = new FormData(document.querySelector("form"))
    fetch('/posts', {
        redirect: 'follow',
        headers: {
            'Accept': 'application/json, text/plain',
            'Content-Type': 'application/json;charset=UTF-8'
        },
        method: 'DELETE',
        body: JSON.stringify(Object.fromEntries(formData))
    }).then((response) => {
        return response.text()
    }).then(data => {
        document.body.innerHTML=data
    })
    return false
}

// https://bulma.io/documentation/form/file/#docsNav
const fileInput = document.querySelector("#file-post-upload input[type=file]");
fileInput.onchange = () => {
    if (fileInput.files.length > 0) {
        const fileName = document.querySelector("#file-post-upload .file-name");
        fileName.textContent = fileInput.files[0].name;
    }
};