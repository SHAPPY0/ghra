const UTILS = {};

function showAlert(type, msg) {
    let modalElm = document.getElementById("alertModal");
    let modal = new bootstrap.Modal(modalElm);
    if (modal) modal.show();
    // let alertModal = document.getElementById("alertModal");
    let alertType = document.getElementById("alertType");
    let alertMsg = document.getElementById("alertMsg");
    if (modalElm) {
        alertType.innerHTML = type;
        alertMsg.innerHTML = msg;
        modalElm.style.display = "block";
    }
}

function openModal(modalId) {
    if (modalId) {
        let modalElm = document.getElementById(modalId);
        let modal = new bootstrap.Modal(modalElm);
        modal.show();
    }
}

function closeModal(modalId) {
    if (!modalId) return;
    let modal = document.getElementById(modalId);
    // let closeButton = modal.querySelector(".close");
    // if (closeButton) closeButton.click();
    let classes = [...modal.classList];
    modal.classList[classes.indexOf("show")] = "hide";
}

function notify(show, status, msg) {
    let notif_elm = document.getElementById("notif_alert");
    let notif_status = document.getElementById("notif_status");
    let notif_msg = document.getElementById("notif_msg");
    if (show) {
        notif_elm.style.display = "block";
        notif_status.innerHTML = status;
        notif_msg.innerHTML = msg;
        setTimeout(() => this.notify(false), 5000);
    } else {
        notif_elm.style.display = "none";
        notif_status.innerHTML = "";
        notif_msg.innerHTML = "";
    }
}

UTILS.formatDate = function(date) {
    let formatted_date = "";
    if (date) {
        date = new Date(date);
        let year = date.getFullYear();
        let month = date.getMonth() + 1;
        month = month > 9 ? month : `0${month}`;
        let day = date.getDay();
        day = day > 9 ? day : `0${day}`;
        formatted_date = `${year} - ${month} - ${day}`;
    }
    return formatted_date
}

UTILS.NewInputTextNode = function(name, id) {
    let input = document.createElement("input");
    input.id = id;
    input.name = name;
    input.placeholder = `Enter ${name}`;
    return input;
}

async function onSignup() {
    let form = document.forms["signupForm"];
    let signupReq = {
        "username": form.username.value,
        "email": form.email.value,
        "password": form.password.value,
        "role": 0,
    }
    const rawResponse = await fetch("/signup", {
        method: "POST",
        headers: {
            'Accept': 'application/json, text/plain, */*',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(signupReq)
    });
    let result = await rawResponse.json();
    if (result && result.status == 200){
        closeModal("signupModal");
        showAlert("Success", result.message);
    } else showAlert("Error", result.message);
}

async function createProject() {
    let form = document.forms["projectForm"];
    let reqData = {
        "name": form.name.value,
        "description": form.description.value,
    }
    const rawResponse = await fetch("/projects", {
        method: "POST",
        headers: {
            'Accept': 'application/json, text/plain, */*',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(reqData)
    });
    let result = await rawResponse.json();
    if (result && result.status == 200){
        closeModal("projectModal");
        // window.location = "/projects";
        showAlert("Success", result.message);
    } else showAlert("Error", result.message);
}

async function addRepository() {
    let form = document.forms["repoForm"];
    let reqData = {
        "projectId": parseInt(form.projectId.value),
        "name": form.name.value,
        "url": form.url.value,
        "branch": form.branch.value,
        "user": form.user.value,
        "token": form.token.value,
        "buildTool": form.buildTool.value,
        "depFilePath": form.depFilePath.value,
        "tags": form.tags.value,
    }
    const rawResponse = await fetch("/repository", {
        method: "POST",
        headers: {
            'Accept': 'application/json, text/plain, */*',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(reqData)
    });
    let result = await rawResponse.json();
    if (result && result.status == 200){
        closeModal("repoModal");
        notify(true, "Success", result.message);
    } else showAlert("Error", result.message);
}

async function commitChanges(e) {
    let form = document.forms["commitForm"];
    let newContent = document.getElementById("content").innerText;
    let cm_err = document.getElementById("cm_err");
    let data = {
        "newContent": document.getElementById("content").innerText,
        "projectId": parseInt(form.projectId.value),
        "repoId": parseInt(form.repoId.value),
        "message": form.commitMessage.value,
        "branch": form.branch.value,
        "sha": form.sha.value,
    }
    if (!data.message) {
        cm_err.style.display = "block";
        return;
    }
    cm_err.style.display = "none";
    const rawResponse = await fetch("/deps", {
        method: "PUT",
        headers: {
            'Accept': 'application/json, text/plain, */*',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    });
    let result = await rawResponse.json();
    if (result && result.status == 200){
        showAlert("Success", result.message);
    } else showAlert("Error", result.message);
}

async function editProject(id) {
    try {
        let modalElm = document.getElementById("projectModal");
        let projectFrom = document.forms["projectForm"];
        const rawResponse = await fetch(`/project/${id}/json`, {
            method: "GET",
            headers: {
                'Accept': 'application/json, text/plain, */*',
                'Content-Type': 'application/json'
            }
        });
        let result = await rawResponse.json();
        let modal = new bootstrap.Modal(modalElm);
        modal.show();
    } catch(ex) {
        console.log(ex)
    }
    
}

async function deleteProject(id) {
    if (!id) return;
    let confm = confirm("Are you sure to delete project?");
    if (!confm) return;
    const rawResponse = await fetch(`/project/${id}`, {
        method: "DELETE",
        headers: {
            'Accept': 'application/json, text/plain, */*',
            'Content-Type': 'application/json'
        }
    });
    let result = await rawResponse.json();
    if (result && result.status == 200){
        showAlert("Success", result.message);
    } else showAlert("Error", result.message);
}

async function editRepo(repoId) {
    let projectId = document.getElementById("projectId").value;
    if (!repoId || !projectId) {
        return;
    }
    const rawResponse = await fetch(`/repository/${projectId}/${repoId}`, {
        method: "GET",
        headers: {
            'Accept': 'application/json, text/plain, */*',
            'Content-Type': 'application/json'
        }
    });
    let result = await rawResponse.json();
    if (result && result.status == 200){
        let data = result.data;
        let modal = new bootstrap.Modal(document.getElementById("repoModal"));
        modal.show();
        let form = document.forms["repoForm"];
        form.url.value = data.Url;
        form.name.value = data.Name;
        form.branch.value = data.Branch;
        form.user.value = data.User;
        form.token.value = data.Token;
        form.buildTool.value = data.BuildTool;
        form.depFilePath.value = data.DepFilePath;
        form.tags.value = data.Tags;
        document.getElementById("updateRepoBtn").style.display = "block";
        document.getElementById("addRepoBtn").style.display = "none";
        document.getElementById("type").innerHTML = "UPDATE";
    } else showAlert("Error", result.message);
}

async function deleteRepo(repoId) {
    if (!window.confirm("Are you sure to delete repository?")) return;
    let projectId = document.getElementById("projectId").value;
    if (!repoId || !projectId) {
        return;
    }
    const rawResponse = await fetch(`/repository/${projectId}/${repoId}`, {
        method: "DELETE",
        headers: {
            'Accept': 'application/json, text/plain, */*',
            'Content-Type': 'application/json'
        }
    });
    let result = await rawResponse.json();
    if (result && result.status == 200) showAlert("Success", result.message);
    else showAlert("Error", result.message);
}

// function closeModal(modalId) {
//     if (modalId) {
//         let modalElm = document.getElementById(modalId);
//         let modal = bootstrap.Modal.getInstance(modalElm);
//         modal.hide();
//     }
// }