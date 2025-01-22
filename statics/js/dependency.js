function dependency() {
    this.properties = {};
    this.dependencies = [];
    this.versionCascade = {
        repositories: [],
    };
}

dependency.prototype.onVersionChange = function(type, version, option) {
    if (type === "property") {
        this.properties[option.name] = version;
    } else if (type === "dependency") {
        let dependency = {
            "groupId": option.groupId,
            "artifactId": option.artifactId,
            "version": version
        }
        let found = false;
        for (let i = 0; i < this.dependencies.length; i++) {
            if (this.dependencies[i].groupId == dependency.groupId && this.dependencies[i].artifactId == dependency.artifactId) {
                this.dependencies[i].version = dependency.version;
                found = true;
                break;
            }
        }
        if (!found) this.dependencies.push(dependency);
    }
}

dependency.prototype.updateChanges = async function() {
    let form = document.forms["commitForm"];
    let cm_err = document.getElementById("cm_err");
    let data = {
        "newContent": {"properties": this.properties, "dependencies": this.dependencies},
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
    const rawResponse = await fetch("http://0.0.0.0:8080/deps", {
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
    } else showAlert("Error", result.error || result.message);
}

dependency.prototype.onRepoSelect = function(repoId) {
    repoId = parseInt(repoId);
    let i = this.versionCascade.repositories.indexOf(repoId); 
    if (i > -1) {
        this.versionCascade.repositories.splice(i, 1);
    } else this.versionCascade.repositories.push(repoId);
    console.log(this.versionCascade.repositories)
}

dependency.prototype.chooseAllRepos = function() {
    // let 
}

dependency.prototype.getVCDeps = async function() {
    let repoIds = this.versionCascade.repositories;
    if (repoIds.length) {
        let data = {
            "repoIds": repoIds,
            "projectId": parseInt(document.getElementById("projectId").value || 0),
        }
        if (!data.repoIds.length) {
            return;
        }
        const rawResponse = await fetch("http://0.0.0.0:8080/vc/deps", {
            method: "POST",
            headers: {
                'Accept': 'application/json, text/plain, */*',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });
        let result = await rawResponse.json();
        if (result && result.status == 200){
            showAlert("Success", result.message);
        } else showAlert("Error", result.error || result.message);
    } else {
        showAlert("Error", "Please choose one or more repositories.");
    }
}

const deps = new dependency();