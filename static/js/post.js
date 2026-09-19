document.addEventListener("DOMContentLoaded", () => {
    const closeMenus = (exceptID = "") => {
        document.querySelectorAll(".owner-menu-dropdown").forEach((menu) => {
            if (menu.id === exceptID) {
                return;
            }

            menu.hidden = true;

            const toggle = document.querySelector(
                `[data-menu-toggle="${menu.id}"]`,
            );

            if (toggle) {
                toggle.setAttribute("aria-expanded", "false");
            }
        });
    };

    const closeReportForms = (exceptID = "") => {
        document.querySelectorAll(".moderation-report-form").forEach((form) => {
            if (form.id === exceptID) {
                return;
            }

            form.hidden = true;
        });
    };

    document.querySelectorAll("[data-menu-toggle]").forEach((button) => {
        button.addEventListener("click", (event) => {
            event.stopPropagation();

            const menuID = button.dataset.menuToggle;
            const menu = document.getElementById(menuID);

            if (!menu) {
                return;
            }

            const shouldOpen = menu.hidden;

            closeMenus(menuID);

            menu.hidden = !shouldOpen;

            button.setAttribute(
                "aria-expanded",
                shouldOpen ? "true" : "false",
            );
        });
    });

    document.querySelectorAll(".owner-menu-dropdown").forEach((menu) => {
        menu.addEventListener("click", (event) => {
            event.stopPropagation();
        });
    });

    document.addEventListener("click", () => {
        closeMenus();
    });

    document.addEventListener("keydown", (event) => {
        if (event.key === "Escape") {
            closeMenus();
            closeReportForms();
        }
    });

    document.querySelectorAll("[data-edit-target]").forEach((button) => {
        button.addEventListener("click", () => {
            const form = document.getElementById(
                button.dataset.editTarget,
            );

            const display = document.getElementById(
                button.dataset.displayTarget,
            );

            if (!form || !display) {
                return;
            }

            closeMenus();
            closeReportForms();

            display.hidden = true;
            form.hidden = false;

            const field = form.querySelector(
                "input[type='text'], textarea",
            );

            if (field) {
                field.focus();
            }
        });
    });

    document.querySelectorAll("[data-cancel-edit]").forEach((button) => {
        button.addEventListener("click", () => {
            const form = document.getElementById(
                button.dataset.cancelEdit,
            );

            const display = document.getElementById(
                button.dataset.displayTarget,
            );

            if (!form || !display) {
                return;
            }

            form.hidden = true;
            display.hidden = false;
        });
    });

    document.querySelectorAll("[data-report-toggle]").forEach((button) => {
        button.addEventListener("click", () => {
            const formID = button.dataset.reportToggle;
            const form = document.getElementById(formID);

            if (!form) {
                return;
            }

            const shouldOpen = form.hidden;

            closeMenus();
            closeReportForms(formID);

            form.hidden = !shouldOpen;

            if (shouldOpen) {
                const textarea = form.querySelector("textarea");

                if (textarea) {
                    textarea.focus();
                }
            }
        });
    });

    document.querySelectorAll("[data-report-close]").forEach((button) => {
        button.addEventListener("click", () => {
            const form = document.getElementById(
                button.dataset.reportClose,
            );

            if (!form) {
                return;
            }

            form.hidden = true;
        });
    });
});