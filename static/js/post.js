// Wait until the initial HTML document has been completely loaded and parsed.
document.addEventListener("DOMContentLoaded", () => {
    // Close all owner dropdown menus except the one identified by exceptID.
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

    // Close all moderation report forms except the one identified by exceptID.
    const closeReportForms = (exceptID = "") => {
        document.querySelectorAll(".moderation-report-form").forEach((form) => {
            if (form.id === exceptID) {
                return;
            }

            form.hidden = true;
        });
    };

    // Attach open and close behavior to every owner menu toggle button.
    document.querySelectorAll("[data-menu-toggle]").forEach((button) => {
        button.addEventListener("click", (event) => {
            // Prevent the document click handler from immediately closing the menu.
            event.stopPropagation();

            const menuID = button.dataset.menuToggle;
            const menu = document.getElementById(menuID);

            if (!menu) {
                return;
            }

            const shouldOpen = menu.hidden;

            // Close any other open owner menus before toggling this one.
            closeMenus(menuID);

            menu.hidden = !shouldOpen;

            // Keep the accessibility state synchronized with the menu visibility.
            button.setAttribute(
                "aria-expanded",
                shouldOpen ? "true" : "false",
            );
        });
    });

    // Prevent clicks inside an owner menu from closing it.
    document.querySelectorAll(".owner-menu-dropdown").forEach((menu) => {
        menu.addEventListener("click", (event) => {
            event.stopPropagation();
        });
    });

    // Close owner menus when the user clicks elsewhere on the page.
    document.addEventListener("click", () => {
        closeMenus();
    });

    // Close open menus and report forms when the Escape key is pressed.
    document.addEventListener("keydown", (event) => {
        if (event.key === "Escape") {
            closeMenus();
            closeReportForms();
        }
    });

    // Attach edit behavior to buttons that target editable content.
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

            // Close unrelated controls before displaying the edit form.
            closeMenus();
            closeReportForms();

            display.hidden = true;
            form.hidden = false;

            // Focus the first editable text field for immediate input.
            const field = form.querySelector(
                "input[type='text'], textarea",
            );

            if (field) {
                field.focus();
            }
        });
    });

    // Attach cancel behavior to edit forms.
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

            // Hide the edit form and restore the normal content display.
            form.hidden = true;
            display.hidden = false;
        });
    });

    // Attach open and close behavior to moderation report forms.
    document.querySelectorAll("[data-report-toggle]").forEach((button) => {
        button.addEventListener("click", () => {
            const formID = button.dataset.reportToggle;
            const form = document.getElementById(formID);

            if (!form) {
                return;
            }

            const shouldOpen = form.hidden;

            // Close other controls before toggling the selected report form.
            closeMenus();
            closeReportForms(formID);

            form.hidden = !shouldOpen;

            // Focus the report textarea when the form is opened.
            if (shouldOpen) {
                const textarea = form.querySelector("textarea");

                if (textarea) {
                    textarea.focus();
                }
            }
        });
    });

    // Attach close behavior to report form close buttons.
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