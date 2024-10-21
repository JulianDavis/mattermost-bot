const rowTemplate = `
<tr>
    <td contenteditable="true">
        <input type="date" id="post_date" name="post_date" value="2018-07-22" min="2018-01-01" max="2018-12-31"/>
    </td>
    <td contenteditable="true">
        <input type="time" id="post_time" name="post_time" value="13:30"/>
    </td>
    <td contenteditable="true">
        <select id="post_channel" name="post_channel">
            <option selected value="Announcements">Announcements</option>
        </select>
    </td>
    <td>
        <textarea id="post_message" name="post_message" class="message form-control" placeholder="Enter your message here" rows="3"></textarea>
    </td>
    <td>
        <span class="table-remove edit">
            <button type="button" class="table-remove btn btn-danger btn-rounder btn-sm my-0">Remove</button>
        </span>
    </td>
</tr>
`;

$(document).ready(function() {
    const $tableId = $('#table');
    const debounceSaveTimeMs = 2500;
    const autoSaveIntervalMs = 5000;
    let dataToSave = {};
    let lastSaveData = {};
    let saveTimeout;
    let lastSaveTime = 0;

    // Function to gather current table data
    function gatherTableData() {
        let tableData = [];
        $tableId.find('tbody tr').each(function() {
            let rowData = {
                date: $(this).find('#post_date').val(),
                time: $(this).find('#post_time').val(),
                channel: $(this).find('#post_channel').val(),
                message: $(this).find('#post_message').val()
            };
            tableData.push(rowData);
        });
        return tableData;
    }

    // Function to save changes with debouncing
    function saveChanges(immediate = false) {
        clearTimeout(saveTimeout);

        // Set a new timeout unless saving immediately
        if (!immediate) {
            saveTimeout = setTimeout(() => {
                performSave();
            }, debounceSaveTimeMs);
        } else {
            performSave();
        }
    }

    // Function to actually perform the save
    function performSave() {
        dataToSave = gatherTableData();

        // Check if data has changed before writing to the server
        if (JSON.stringify(dataToSave) !== JSON.stringify(lastSaveData)) {
            console.log("Sending data to the server:", dataToSave);
            lastSaveData = JSON.parse(JSON.stringify(dataToSave)); // Update last saved data
            lastSaveTime = Date.now(); // Update last save time
            writeDataToServer(dataToSave);
        }
    }

    // Function to write data to the server
    function writeDataToServer(data) {
        console.log("Trying to POST data to be saved");
        fetch('/save', {
            method: 'POST',
            headers: {
                'Content-Type': 'application/json',
            },
            body: JSON.stringify(data),
        })
        .then(response => {
            if (!response.ok) {
                throw new Error('Network response was not ok');
            }
            return response.text();
        })
        .then(data => console.log(data))
        .catch((error) => console.error('Error saving data:', error));
    }

    // Function to load the table data
    function loadTableData() {
        console.log("Loading saved data...");
        $.getJSON('/data', function(data) {
            // Clear existing rows
            $tableId.find('tbody').empty();

            console.log("data: ", data)
            // Populate the table with the data
            data.forEach(row => {
                let $newRow = $(rowTemplate);

                $newRow.find('#post_date').val(row.date);
                $newRow.find('#post_time').val(row.time);
                $newRow.find('#post_channel').val(row.channel);
                $newRow.find('#post_message').val(row.message);

                $tableId.find("tbody").append($newRow);
            });
        })
        .fail(function() {
            console.error('Error loading table data.');
        });
    }

    // Start auto-save at intervals, triggering save if enough time has passed
    function startAutoSave() {
        setInterval(function() {
            let now = Date.now();
            // If last save was more than debounceSaveTimeMs ago, force a save
            if (now - lastSaveTime > debounceSaveTimeMs) {
                saveChanges(true); // Trigger immediate save
            }
        }, autoSaveIntervalMs);
    }

    // Load the data and start auto-saving
    loadTableData();
    startAutoSave();

    // Add new row
    $(".table-add").on("click", function() {
        let newRowHtml = $(rowTemplate);
        let $newRow = $(newRowHtml); // Create a jQuery object

        // Populate with current date and time
        let currentDate = new Date();
        let formattedDate = currentDate.toISOString().split('T')[0];
        let formattedTime = currentDate.toTimeString().split(' ')[0].substring(0, 5);

        $newRow.find('input[type="date"]').val(formattedDate);
        $newRow.find('input[type="time"]').val(formattedTime);
        $newRow.find('.message').val(''); // Clear input for message

        // Append the new row
        $tableId.find("tbody").append($newRow);
    });

    // Remove rows
    $tableId.on("click", ".table-remove", function() {
        $(this).parents("tr").detach();
    });

    // Capture input change events
    $tableId.on('input change', 'input[type="date"], input[type="time"], select, .message', function() {
        saveChanges(); // Call saveChanges to start the debouncing process
    });

    // Handle focus-out events to trigger saving
    $tableId.on('blur', 'input[type="date"], input[type="time"], select, .message', function() {
        saveChanges(); // Save changes when focus is lost
    });

    // Handle data saving before user closes the page
    $(window).on('beforeunload', function() {
        saveChanges(true); // Force immediate save before leaving
    });
});
