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
    let chooseAll = document.getElementById("chooseAllRepos");
    let form = document.forms["repoListForm"];
    if (chooseAll.checked) {
        form.repositories.forEach(repo => {
            let repoId = parseInt(repo.value);
            if (this.versionCascade.repositories.indexOf(repoId) == -1) {
                this.versionCascade.repositories.push(parseInt(repo.value));
                repo.checked = true;
            }
        });
    } else {
        this.versionCascade.repositories = [];
        form.repositories.forEach(repo => {
            repo.checked = false;
        })
    }
    console.log(this.versionCascade.repositories);
}

dependency.prototype.vcBack = function() {
    document.getElementById("vcCancel").style.display = "block";
    document.getElementById("vcNext").style.display = "block";
    document.getElementById("vcBack").style.display = "none";
    document.getElementById("vcPushChanges").style.display = "none";
    openTab(0);
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
        const rawResponse = await fetch("/vc/deps", {
            method: "POST",
            headers: {
                'Accept': 'application/json, text/plain, */*',
                'Content-Type': 'application/json'
            },
            body: JSON.stringify(data)
        });
        let result = await rawResponse.json();
        if (result && result.status == 200){
            document.getElementById("vcCancel").style.display = "none";
            document.getElementById("vcNext").style.display = "none";
            document.getElementById("vcBack").style.display = "block";
            document.getElementById("vcPushChanges").style.display = "block";
            bindVCDeps(result);
            openTab(1);
        } else showAlert("Error", result.error || result.message);
    } else {
        showAlert("Error", "Please choose one or more repositories.");
    }
}

function bindVCDeps(result) {
    let form = document.forms["vcDepsForm"];
    let propertiesDom = document.getElementById("properties");
    let dependencies = document.getElementById("dependencies");
    let { Properties, Dependencies } = result.data;
    let propList = "";
    for (let prop in Properties) {
        propList += `<li>
                    <pre><span id="${prop}">${prop}:</span>
                    <input type="text" value="${Properties[prop]}" onchange="deps.onVersionChange('property', this.value, {'name': '${prop}'})" /></pre>
                </li>`;
    }
    if (propList) propertiesDom.innerHTML = propList;
    else propertiesDom.innerHTML = `<div>No common properties found</div>`;

    let depsList = "";
    for (let i = 0; i < Dependencies.length; i++) {
        let dep = Dependencies[i];
        depsList += `<div class="card">
                            <pre style="font-size: 14px;">${dep.GroupID}</pre>
                            <pre class="inner_child">&nbsp;&nbsp;&nbsp;&nbsp;${dep.ArtifactID}</pre>
                            <p class="inner_child">&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;&nbsp;
                                <input type="text" value="${dep.Version}" onchange="deps.onVersionChange('dependency', this.value, {'groupId': '${dep.GroupID}', 'artifactId': '${dep.ArtifactID}'})" />
                            </p>
                        </div>`;
    }
    if (depsList) dependencies.innerHTML = depsList;
    else dependencies.innerHTML = `<div>No common dependencies found</div>`;

    form.projectId.value = result.data.ProjectId;
    form.repoIds.value = result.data.RepoIds;
}

function openTab(tabNo) {
    // Get all tabs and tab panes
    var tabs = document.querySelectorAll("#myTab .nav-link");
    var tabPanes = document.querySelectorAll("#myTabContent .tab-pane");

    // Find the active tab
    var activeIndex = 0;
    tabs.forEach((tab, index) => {
        if (tab.classList.contains("active")) {
            activeIndex = index;
        }
    });

    // Remove the active class from the current tab and tab pane
    tabs[activeIndex].classList.remove("active");
    tabs[activeIndex].ariaSelected = "false";
    tabPanes[activeIndex].classList.remove("show", "active");

    // Calculate the index of the next tab (wrap around if at the end)
    var nextIndex = tabNo;

    // Add the active class to the next tab and tab pane
    tabs[nextIndex].classList.add("active");
    tabs[nextIndex].ariaSelected = "true";
    tabPanes[nextIndex].classList.add("show", "active");
}

dependency.prototype.vcPushChanges = async function() {
    let form = document.forms["vcDepsForm"];
    let commitMsg = document.getElementById("message").value;
    let repoIds = [];
    let data = [];
    form.repoIds.value.split(",").forEach(id => repoIds.push(parseInt(id)));
    if (!commitMsg) {
        alert("Please enter commit message")
        return;
    }

    for (let i = 0; i < repoIds.length; i++) {
        data.push({
            "newContent": {"properties": this.properties, "dependencies": this.dependencies},
            "projectId": parseInt(form.projectId.value),
            "repoId": repoIds[i],
            "message": commitMsg
        });
    }
    const rawResponse = await fetch("/vc/deps", {
        method: "PUT",
        headers: {
            'Accept': 'application/json, text/plain, */*',
            'Content-Type': 'application/json'
        },
        body: JSON.stringify(data)
    });
    let result = await rawResponse.json();
    if (result && result.status == 200){
        msg = result.message + "\n\n" + JSON.stringify(result.data);
        closeModal("versionCascadeModal");
        showAlert("Success", msg);
    } else {
        msg = (result.error || result.message) + "\n\n" + JSON.stringify(result.data);
        showAlert("Error", msg);
    } 
}

const deps = new dependency();