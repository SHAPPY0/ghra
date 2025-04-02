function UTILS1(){
    this.modal = null;
};

UTILS1.prototype.openModal = function(modalId) {
    let modalElm = document.getElementById(modalId)
    let modal = new bootstrap.Modal(modalElm, {
        backdrop: "static",
        keyboard: false
    }); 
    modalElm.addEventListener("click", function (event) {
        if (event.target.matches("[data-bs-dismiss='modal']") && !event.target.classList.contains("btn-close")) {
            event.preventDefault();
            event.stopPropagation();
        }
    });

    //Check if form exist in modal then reset it
    let form = modalElm.querySelector("form");
    if (form) {
        form.reset();
        let select = form.querySelector("select");
        if (select) {
            for (let i = select.options.length - 1; i >= 0; i--) {
                if (select.options[i].value !== "") {
                    select.remove(i);
                }
            }
        }
    }
    modal.show();
    this.modal = modal;
}

UTILS1.prototype.closeModal = function(modalId) {
    if (!modalId) return;
    let modal = document.getElementById(modalId);
    let classes = [...modal.classList];
    console.log(this.modal);
    if (this.modal) this.modal.hide();
}

const utils = new UTILS1();