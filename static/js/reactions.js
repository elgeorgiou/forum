document.addEventListener("DOMContentLoaded", () => {
    const reactionForms = document.querySelectorAll(".reaction-form");

    reactionForms.forEach((form) => {
        form.addEventListener("submit", async (event) => {
            event.preventDefault();

            const button = form.querySelector(".reaction-button");

            if (!button || button.disabled) {
                return;
            }

            button.disabled = true;

            try {
                const response = await fetch(form.action, {
                    method: "POST",
                    body: new FormData(form),
                    headers: {
                        "X-Requested-With": "fetch",
                    },
                });

                if (!response.ok) {
                    throw new Error("Reaction request failed");
                }

                const result = await response.json();

                const reactionContainer = form.closest(
                    ".reaction-group, .comment-reactions"
                );

                if (!reactionContainer) {
                    return;
                }

                const likeButton = reactionContainer.querySelector(
                    ".reaction-like"
                );

                const dislikeButton = reactionContainer.querySelector(
                    ".reaction-dislike"
                );

                if (!likeButton || !dislikeButton) {
                    return;
                }

                likeButton.classList.toggle(
                    "is-active",
                    result.reaction === 1
                );

                dislikeButton.classList.toggle(
                    "is-active",
                    result.reaction === -1
                );

                const likeCount = likeButton.querySelector(
                    ".reaction-count"
                );

                const dislikeCount = dislikeButton.querySelector(
                    ".reaction-count"
                );

                if (likeCount) {
                    likeCount.textContent = result.likes;
                }

                if (dislikeCount) {
                    dislikeCount.textContent = result.dislikes;
                }
            } catch (error) {
                console.error(error);
            } finally {
                button.disabled = false;
            }
        });
    });
});