// Client-side interactions for TaskFlow

// Modal toggles and form populator
function openAddTaskModal() {
    const modal = document.getElementById('taskModal');
    const modalTitle = document.getElementById('modalTitle');
    const taskForm = document.getElementById('taskForm');
    const taskID = document.getElementById('taskID');
    const taskTitle = document.getElementById('taskTitle');
    const taskDescription = document.getElementById('taskDescription');
    const taskCategory = document.getElementById('taskCategory');
    const taskSource = document.getElementById('taskSource');
    const taskStatus = document.getElementById('taskStatus');
    const taskDueDate = document.getElementById('taskDueDate');
    const taskClientID = document.getElementById('taskClientID');

    if (!modal) return;

    modalTitle.textContent = "Tambah Tugas Baru";
    taskForm.action = "/tasks/create";
    taskID.value = "";
    taskForm.reset();
    
    // Set default selections
    if (window.jQuery) {
        $(taskCategory).val("Bugs").trigger('change');
        $(taskSource).val("WA Supp").trigger('change');
        $(taskStatus).val("Pending").trigger('change');
        if (taskClientID) $(taskClientID).val("").trigger('change');
    } else {
        taskCategory.value = "Bugs";
        taskSource.value = "WA Supp";
        taskStatus.value = "Pending";
        if (taskClientID) taskClientID.value = "";
    }
    
    modal.style.display = "flex";
}

function openEditTaskModal(btn) {
    const modal = document.getElementById('taskModal');
    const modalTitle = document.getElementById('modalTitle');
    const taskForm = document.getElementById('taskForm');
    const taskID = document.getElementById('taskID');
    const taskTitle = document.getElementById('taskTitle');
    const taskDescription = document.getElementById('taskDescription');
    const taskCategory = document.getElementById('taskCategory');
    const taskSource = document.getElementById('taskSource');
    const taskStatus = document.getElementById('taskStatus');
    const taskDueDate = document.getElementById('taskDueDate');
    const taskClientID = document.getElementById('taskClientID');

    if (!modal) return;

    modalTitle.textContent = "Edit Tugas";
    taskForm.action = "/tasks/update";
    
    // Populate form fields using data attributes from button
    taskID.value = btn.getAttribute('data-id') || "";
    taskTitle.value = btn.getAttribute('data-title') || "";
    taskDescription.value = btn.getAttribute('data-description') || "";
    taskDueDate.value = btn.getAttribute('data-due-date') || "";
    
    if (window.jQuery) {
        $(taskCategory).val(btn.getAttribute('data-category') || "Bugs").trigger('change');
        $(taskSource).val(btn.getAttribute('data-source') || "WA Supp").trigger('change');
        $(taskStatus).val(btn.getAttribute('data-status') || "Pending").trigger('change');
        if (taskClientID) $(taskClientID).val(btn.getAttribute('data-client-id') || "").trigger('change');
    } else {
        taskCategory.value = btn.getAttribute('data-category') || "Bugs";
        taskSource.value = btn.getAttribute('data-source') || "WA Supp";
        taskStatus.value = btn.getAttribute('data-status') || "Pending";
        if (taskClientID) taskClientID.value = btn.getAttribute('data-client-id') || "";
    }
    
    modal.style.display = "flex";
}

function closeTaskModal() {
    const modal = document.getElementById('taskModal');
    if (modal) {
        modal.style.display = "none";
    }
}

// Global modal background-clicking listener to dismiss
window.addEventListener('click', function(event) {
    const modal = document.getElementById('taskModal');
    if (event.target === modal) {
        closeTaskModal();
    }
});

// Auto-close alert notifications after 5 seconds
document.addEventListener("DOMContentLoaded", function() {
    const alerts = document.querySelectorAll('.alert');
    alerts.forEach(alert => {
        setTimeout(() => {
            alert.style.transition = "opacity 0.5s ease";
            alert.style.opacity = "0";
            setTimeout(() => {
                alert.remove();
            }, 500);
        }, 5000);
    });
});
