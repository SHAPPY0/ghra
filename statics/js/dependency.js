function dependency() {
    this.properties = {};
    this.parent = {};
    this.dependencies = [];
    this.versionCascade = {
        repositories: [],
        selectedTab: "v-pills-cr-tab",
    };
}

dependency.prototype.maven = {
    newProperties: [],
    newDependencies: [],
    addProperty: function() {
        this.newProperties.push({"name": "", "version": ""});
        this.renderNewPropForm();
    },
    addDependency: function() {
        this.newDependencies.push({"groupId": "", "artifactId": "", "version": ""});
        this.renderNewDepsForm();
    },
    renderNewPropForm: function() {
        let propsElm = document.getElementById("properties");
        this.resetPropsForm();
        for(let i = 0; i < this.newProperties.length; i++) {
            let li = document.createElement("li");
            li.className = "newProp";
            let nameInput = UTILS.NewInputTextNode("name", "name");
            nameInput.setAttribute("onchange", `deps.maven.onPropertyChange(${i}, 'name', this.value)`)
            let colon = document.createElement("span");
            colon.innerHTML = "&nbsp;:&nbsp;";
            let versionInput = UTILS.NewInputTextNode("version", "version");
            versionInput.setAttribute("onchange", `deps.maven.onPropertyChange(${i}, 'version', this.value)`)
            let removeIcon = document.createElement("i");
            removeIcon.className = "fa fa-remove propRemove";
            removeIcon.setAttribute("onclick", `deps.maven.removeProperty('${i}')`);
            li.appendChild(nameInput);
            li.appendChild(colon);
            li.appendChild(versionInput);
            li.appendChild(removeIcon);
            propsElm.appendChild(li);
        }
    },
    renderNewDepsForm: function() {
        let depsElm = document.getElementById("deps");
        this.resetDepsForm();
        for (let i = 0; i < this.newDependencies.length; i++) {
            let div = document.createElement("div");
            div.className = "card newDep";
            for (let k in this.newDependencies[i]) {
                let pre = document.createElement("pre");
                let span = document.createElement("span");
                let b = document.createElement("b");
                b.innerHTML = `${k}: `;
                span.appendChild(b);
                pre.appendChild(span);
                let input = UTILS.NewInputTextNode(k, k);
                input.setAttribute("onchange", `deps.maven.onDependencyChange(${i}, '${k}', this.value)`)
                pre.appendChild(input);
                div.appendChild(pre);
            }
            depsElm.appendChild(div);
        }
    },
    onPropertyChange: function(index, key, value) {
        this.newProperties[index][key] = value;
    },
    onDependencyChange: function(index, key, value) {
        this.newDependencies[index][key] = value;
    },
    removeProperty: function(index) {
        this.newProperties.splice(index, 1);
        this.renderNewPropForm();
    },
    removeDependency: function(index) {
        this.newDependencies.splice(index, 1);
        this.renderNewDepsForm();
    },
    resetPropsForm: function() {
        let propsElm = document.getElementById("properties");
        let newProps = propsElm.getElementsByClassName("newProp");
        let newPropList = [...newProps];
        for (let i = 0; i < newPropList.length; i++) {
            newProps[0].remove();
        }
    },
    resetDepsForm: function() {
        let depElm = document.getElementById("deps");
        let newDep = depElm.getElementsByClassName("newDep");
        let newDepsList = [...newDep];
        for (let i = 0; i < newDepsList.length; i++) {
            newDep[0].remove();
        }

    }
};

dependency.prototype.onVersionChange = function(type, version, option) {
    if (type === "property") {
        this.properties[option.name] = version;
    } else if (type === "parent") {
        this.parent = {
            "groupId": option.groupId,
            "artifactId": option.artifactId,
            "version": version
        }
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
        "content": {
            "properties": this.properties, 
            "dependencies": this.dependencies,
            "parent": this.parent,
            "newProperties": this.maven.newProperties,
            "newDependencies": this.maven.newDependencies
        },
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
    // let form = document.forms["repoListForm"];
    // form.repositories = Array.isArray(form.repositories) ? form.repositories : [form.repositories];
    let repos = document.getElementsByName("repository");
    if (chooseAll.checked) {
        repos.forEach(repo => { 
            let repoId = parseInt(repo.value);
            if (this.versionCascade.repositories.indexOf(repoId) == -1) {
                this.versionCascade.repositories.push(parseInt(repo.value));
                repo.checked = true;
            }
        });
    } else {
        this.versionCascade.repositories = [];
        repos.forEach(repo => {
            repo.checked = false;
        })
    }
}

dependency.prototype.vcBack = function() {
    document.getElementById("vcCancel").style.display = "block";
    document.getElementById("vcNext").style.display = "block";
    document.getElementById("vcBack").style.display = "none";
    document.getElementById("vcPushChanges").style.display = "none";
    openTab(0);
}

dependency.prototype.checkReposSelected = function() {
    let repoIds = this.versionCascade.repositories;
    if (!repoIds.length) {
        showAlert("Error", "Please choose one or more repositories.");
        return false;
    }
    return true;
}

dependency.prototype.getVCDeps = async function() {
    let repoIds = this.versionCascade.repositories;

    if (deps.checkReposSelected()) {
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
            bindVCDeps(result);
            // openTab(1);
        } else showAlert("Error", result.error || result.message);
    }
}

function bindVCDeps(result) {
    let form = document.forms["vcDepsForm"];
    let parentDom = document.getElementById("parent");
    let propertiesDom = document.getElementById("properties");
    let dependencies = document.getElementById("dependencies");
    let { Parent, Properties, Dependencies } = result.data;

    let parent = "";
    if (Parent && Object.keys(Parent).length) {
        parent += `<div class="card">
                            <pre><b>GroupID:</b> ${Parent.GroupID}</pre>
                            <pre><b>ArtifactID:</b> ${Parent.ArtifactID}</pre>
                            <pre><b>Version:</b> ${!Parent.Version ? `<input type="text" value="" onchange="deps.onVersionChange('parent', this.value, {'groupId': '${Parent.GroupID}', 'artifactId': '${Parent.ArtifactID}'})" />` : 
                                `<input type="text" value="${Parent.Version}" onchange="deps.onVersionChange('parent', this.value, {'groupId': '${Parent.GroupID}', 'artifactId': '${Parent.ArtifactID}'})" />`
                            }</pre>
                        </div>`
    }

    if (parent) parentDom.innerHTML = parent;
    else parentDom.innerHTML = `<div>No common parent found</div>`;

    let propList = "";
    for (let prop in Properties) {
        propList += `<li>
                    <pre id="${prop}">${prop}:<input type="text" value="${Properties[prop]}" onchange="deps.onVersionChange('property', this.value, {'name': '${prop}'})" /></pre>
                </li>`;
    }
    if (propList) propertiesDom.innerHTML = propList;
    else propertiesDom.innerHTML = `<div>No common properties found</div>`;

    let depsList = "";
    for (let i = 0; i < Dependencies.length; i++) {
        let dep = Dependencies[i];
        depsList += `<div class="card">
                            <pre><b>GroupID:</b> ${dep.GroupID}</pre>
                            <pre><b>ArtifactID:</b> ${dep.ArtifactID}</pre>
                            <pre><b>Version:</b> ${!dep.Version ? `<input type="text" value="" onchange="deps.onVersionChange('dependency', this.value, {'groupId': '${dep.GroupID}', 'artifactId': '${dep.ArtifactID}'})" />` : 
                                `<input type="text" value="${dep.Version}" onchange="deps.onVersionChange('dependency', this.value, {'groupId': '${dep.GroupID}', 'artifactId': '${dep.ArtifactID}'})" />`
                            }</pre>
                        </div>`;
    }
    if (depsList) dependencies.innerHTML = depsList;
    else dependencies.innerHTML = `<div>No common dependencies found</div>`;

    form.projectId.value = result.data.ProjectId;
    form.repoIds.value = result.data.RepoIds;
}

function openTab(tabNo) {
    // Get all tabs and tab panes
    var tabs = document.querySelectorAll("#cascadeTabs .nav-link");
    var tabPanes = document.querySelectorAll("#cascadeTabPanel .tab-pane");

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
    let commitMsg = document.getElementById("commitMessage").value;
    let repoIds = this.versionCascade.repositories;;
    let data = [];
    // form.repoIds.value.split(",").forEach(id => repoIds.push(parseInt(id)));
    if (!commitMsg) {
        alert("Please enter commit message")
        return;
    }

    for (let i = 0; i < repoIds.length; i++) {
        data.push({
            "content": {
                "parent": this.parent,
                "properties": this.properties, 
                "dependencies": this.dependencies, 
                "newProperties": [], 
                "newDependencies": []
            },
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
        // closeModal("updateCascadeModal");
        showAlert("Success", msg);
    } else {
        msg = (result.error || result.message) + "\n\n" + JSON.stringify(result.data);
        showAlert("Error", msg);
    } 
}

document.addEventListener("DOMContentLoaded", function() {
    const tabLinks = document.querySelectorAll('a[data-toggle="pill"]');
    tabLinks.forEach(tab => {
        tab.addEventListener("click", function (event) {
            event.preventDefault();
            // Remove active class from all tabs
            tabLinks.forEach(t => t.classList.remove("active"));
            this.classList.add("active");
            // Get target tab pane ID from href
            const targetTabPane = document.querySelector(this.getAttribute("href"));
            // Hide all tab panes
            document.querySelectorAll(".tab-pane").forEach(pane => pane.classList.remove("show", "active"));
            // Show the clicked tab pane
            targetTabPane.classList.add("show", "active");
            deps.navPillChange(this.id);
        });
    });

    
});

dependency.prototype.navPillChange = function(tabId) {
    if (this.versionCascade.selectedTab != tabId) {
        this.versionCascade.selectedTab = tabId;
        if (tabId === "v-pills-uv-tab") deps.getVCDeps();
        else if (tabId === "v-pills-pc-tab") deps.checkReposSelected();
    }
}

dependency.prototype.goToTab = function(tabId) {
    if (tabId) {
        document.getElementById(tabId).click();
    }
}

const deps = new dependency();