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