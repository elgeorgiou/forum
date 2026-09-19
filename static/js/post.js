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
});