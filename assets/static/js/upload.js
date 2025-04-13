document.addEventListener("DOMContentLoaded", function () {
  // ======== UPLOAD TAB FUNCTIONALITY ========

  // Store upload URLs and task ID
  let uploadData = null;
  let transactionUploaded = false;
  let bankStatementUploaded = false;

  // DOM elements
  const form = document.getElementById("reconciliationForm");
  const submitButton = document.getElementById("submitBtn");
  const transactionFileInput = document.getElementById("transactionFile");
  const bankStatementFileInput = document.getElementById("bankStatementFile");
  const uploadTransactionBtn = document.getElementById("uploadTransactionBtn");
  const uploadBankStatementBtn = document.getElementById(
    "uploadBankStatementBtn",
  );

  // Progress and status elements
  const transactionProgress = document.getElementById("transactionProgress");
  const bankStatementProgress = document.getElementById(
    "bankStatementProgress",
  );
  const transactionSuccess = document.getElementById("transactionSuccess");
  const bankStatementSuccess = document.getElementById("bankStatementSuccess");
  const transactionError = document.getElementById("transactionError");
  const bankStatementError = document.getElementById("bankStatementError");

  // Modal elements
  const notificationModal = new bootstrap.Modal(
    document.getElementById("notificationModal"),
  );
  const modalTitle = document.getElementById("modalTitle");
  const modalBody = document.getElementById("modalBody");

  // Tab Navigation
  const summariesTab = document.getElementById("summaries-tab");
  summariesTab.addEventListener("shown.bs.tab", function (e) {
    loadSummaries();
  });

  // Initially fetch upload URLs
  fetchUploadUrls();

  // Handle transaction file upload button
  uploadTransactionBtn.addEventListener("click", async function () {
    if (!uploadData || !uploadData.transactionUrl) {
      showModal(
        "Error",
        "Upload URL not available. Please try refreshing the page.",
      );
      return;
    }

    const filename = await uploadFile(
      transactionFileInput,
      transactionProgress,
      transactionSuccess,
      transactionError,
      uploadData.transactionUrl, // Pass the URL directly instead of type
    );

    if (filename) {
      transactionUploaded = true;
      checkUploadStatus();
    }
  });

  // Handle bank statement file upload button
  uploadBankStatementBtn.addEventListener("click", async function () {
    if (!uploadData || !uploadData.bankStatementUrl) {
      showModal(
        "Error",
        "Upload URL not available. Please try refreshing the page.",
      );
      return;
    }

    const filename = await uploadFile(
      bankStatementFileInput,
      bankStatementProgress,
      bankStatementSuccess,
      bankStatementError,
      uploadData.bankStatementUrl, // Pass the URL directly instead of type
    );

    if (filename) {
      bankStatementUploaded = true;
      checkUploadStatus();
    }
  });

  // Form submission handler
  form.addEventListener("submit", function (e) {
    e.preventDefault();

    if (!uploadData || !transactionUploaded || !bankStatementUploaded) {
      showModal("Error", "Please upload both files first.");
      return;
    }

    const bankName = document.getElementById("bankName").value.trim();
    if (!bankName) {
      showModal("Error", "Please enter a bank name.");
      return;
    }

    // Collect form data
    const startDateInput = document.getElementById("startDate").value;
    const endDateInput = document.getElementById("endDate").value;

    // Create request data object - fix the taskID property name to match API response
    const requestData = {
      taskID: uploadData.taskID, // Changed from taskId to taskID to match server response
      bankName: bankName,
    };

    // Format dates in Go's time format if provided
    if (startDateInput) {
      // Set time to beginning of day (00:00:00) and format in ISO8601
      const startDate = new Date(startDateInput);
      startDate.setHours(0, 0, 0, 0);
      requestData.startDate = startDate.toISOString();
    }

    if (endDateInput) {
      // Set time to end of day (23:59:59) and format in ISO8601
      const endDate = new Date(endDateInput);
      endDate.setHours(23, 59, 59, 999);
      requestData.endDate = endDate.toISOString();
    }

    // Log the formatted data for debugging
    console.log("Submitting reconciliation with data:", requestData);

    // Submit to initiate compilation
    submitButton.disabled = true;
    submitButton.innerHTML =
      '<span class="spinner-border spinner-border-sm" role="status" aria-hidden="true"></span> Processing...';

    fetch("/api/reconciliation", {
      method: "POST",
      headers: {
        "Content-Type": "application/json",
      },
      body: JSON.stringify(requestData),
    })
      .then((response) => {
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        return response.json();
      })
      .then((data) => {
        showModal(
          "Success",
          "Compilation process has been initiated successfully!",
        );
        form.reset();
        transactionUploaded = false;
        bankStatementUploaded = false;
        hideElement(transactionSuccess);
        hideElement(bankStatementSuccess);
        submitButton.disabled = true;

        // Get new upload URLs for next submission
        fetchUploadUrls();

        // Switch to summaries tab to show results
        setTimeout(() => {
          const summariesTab = document.getElementById("summaries-tab");
          const tab = new bootstrap.Tab(summariesTab);
          tab.show();
        }, 1500);
      })
      .catch((error) => {
        showModal("Error", "Failed to initiate compilation: " + error.message);
      })
      .finally(() => {
        submitButton.disabled = false;
        submitButton.textContent = "Start Compilation";
      });
  });

  // Function to fetch upload URLs
  function fetchUploadUrls() {
    fetch("/api/reconciliation/upload")
      .then((response) => {
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        return response.json();
      })
      .then((data) => {
        uploadData = data;
        console.log("Upload URLs received:", uploadData);
      })
      .catch((error) => {
        console.error("Error fetching upload URLs:", error);
        showModal("Error", "Failed to get upload URLs: " + error.message);
      });
  }

  // Simplified uploadFile function that uses direct POST without URL parsing
  async function uploadFile(
    fileInput,
    progressBar,
    successElement,
    errorElement,
    uploadUrl,
  ) {
    // Reset UI elements
    progressBar.classList.remove("d-none");
    successElement.classList.add("d-none");
    errorElement.classList.add("d-none");

    const file = fileInput.files[0];
    if (!file) {
      errorElement.textContent = "Please select a file first";
      errorElement.classList.remove("d-none");
      progressBar.classList.add("d-none");
      return null;
    }

    try {
      console.log("Starting upload to URL:", uploadUrl);

      // Use the URL directly without parsing
      const uploadResponse = await fetch(uploadUrl, {
        method: "POST",
        body: file,
        headers: {
          "Content-Type": file.type || "text/csv",
        },
      });

      if (!uploadResponse.ok) {
        const errorText = await uploadResponse.text();
        throw new Error(
          `Upload failed with status: ${uploadResponse.status} - ${errorText}`,
        );
      }

      // Show success message
      successElement.textContent = "File uploaded successfully!";
      successElement.classList.remove("d-none");
      progressBar.querySelector(".progress-bar").style.width = "100%";

      // Extract filename from URL for form submission (simplest approach)
      const urlParts = uploadUrl.split("/");
      return urlParts[urlParts.length - 1]; // Just return the last part of the URL path
    } catch (error) {
      console.error("Upload failed:", error);

      // Add detailed error information
      let errorMsg = error.message;
      if (
        error.name === "TypeError" &&
        error.message.includes("Failed to fetch")
      ) {
        errorMsg =
          "Network error: Check that the bucket URL is accessible from your browser";
      }

      errorElement.textContent = errorMsg;
      errorElement.classList.remove("d-none");
      return null;
    } finally {
      progressBar.classList.add("d-none");
    }
  }

  // Function to check if both files are uploaded and enable submit button
  function checkUploadStatus() {
    submitButton.disabled = !(transactionUploaded && bankStatementUploaded);
  }

  // ======== SUMMARIES TAB FUNCTIONALITY ========

  // DOM elements for summaries
  const refreshButton = document.getElementById("refreshSummaries");
  const summariesLoading = document.getElementById("summariesLoading");
  const summariesTable = document.getElementById("summariesTable");
  const noSummaries = document.getElementById("noSummaries");
  const summariesTableBody = document.getElementById("summariesTableBody");
  const summariesPagination = document.getElementById("summariesPagination");
  const summariesPrevBtn = document.getElementById("summariesPrevBtn");
  const summariesNextBtn = document.getElementById("summariesNextBtn");
  const summariesCurrentRange = document.getElementById(
    "summariesCurrentRange",
  );
  const summariesTotalCount = document.getElementById("summariesTotalCount");

  // DOM elements for details modal
  const detailsModal = new bootstrap.Modal(
    document.getElementById("detailsModal"),
  );
  const detailsModalTitle = document.getElementById("detailsModalTitle");
  const detailsLoading = document.getElementById("detailsLoading");
  const detailsTable = document.getElementById("detailsTable");
  const noDetails = document.getElementById("noDetails");
  const detailsTableHead = document.getElementById("detailsTableHead");
  const detailsTableBody = document.getElementById("detailsTableBody");
  const detailsPagination = document.getElementById("detailsPagination");
  const detailsPrevBtn = document.getElementById("detailsPrevBtn");
  const detailsNextBtn = document.getElementById("detailsNextBtn");
  const detailsCurrentRange = document.getElementById("detailsCurrentRange");
  const detailsTotalCount = document.getElementById("detailsTotalCount");

  // Pagination state for summaries
  const summaryState = {
    limit: 10,
    offset: 0,
    totalCount: 0,
  };

  // Pagination state for details
  const detailState = {
    taskId: null,
    type: null, // 'bank' or 'transaction'
    limit: 10,
    offset: 0,
    totalCount: 0,
  };

  // Handle refresh button
  refreshButton.addEventListener("click", function () {
    loadSummaries();
  });

  // Handle summary pagination
  summariesPrevBtn.addEventListener("click", function (e) {
    e.preventDefault();
    if (summaryState.offset > 0) {
      summaryState.offset -= summaryState.limit;
      if (summaryState.offset < 0) summaryState.offset = 0;
      loadSummaries();
    }
  });

  summariesNextBtn.addEventListener("click", function (e) {
    e.preventDefault();
    if (summaryState.offset + summaryState.limit < summaryState.totalCount) {
      summaryState.offset += summaryState.limit;
      loadSummaries();
    }
  });

  // Handle details pagination
  detailsPrevBtn.addEventListener("click", function (e) {
    e.preventDefault();
    if (detailState.offset > 0) {
      detailState.offset -= detailState.limit;
      if (detailState.offset < 0) detailState.offset = 0;
      loadDetails();
    }
  });

  detailsNextBtn.addEventListener("click", function (e) {
    e.preventDefault();
    if (detailState.offset + detailState.limit < detailState.totalCount) {
      detailState.offset += detailState.limit;
      loadDetails();
    }
  });

  // Function to load reconciliation summaries
  function loadSummaries() {
    // Show loading state
    showElement(summariesLoading);
    hideElement(summariesTable);
    hideElement(noSummaries);
    hideElement(summariesPagination);

    fetch(
      `/api/reconciliation/summary/list?limit=${summaryState.limit}&offset=${summaryState.offset}`,
    )
      .then((response) => {
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        return response.json();
      })
      .then((data) => {
        const summaries = data.data;
        summaryState.totalCount = data.totalCount;

        // Update table
        displaySummaries(summaries);

        // Update pagination info
        updatePaginationInfo(
          summaryState,
          summariesCurrentRange,
          summariesTotalCount,
          summariesPrevBtn,
          summariesNextBtn,
        );

        // Show appropriate UI elements
        hideElement(summariesLoading);
        if (summaries.length === 0) {
          showElement(noSummaries);
          hideElement(summariesTable);
          hideElement(summariesPagination);
        } else {
          hideElement(noSummaries);
          showElement(summariesTable);
          showElement(summariesPagination);
        }
      })
      .catch((error) => {
        console.error("Error fetching summaries:", error);
        hideElement(summariesLoading);
        showElement(noSummaries);
        noSummaries.textContent =
          "Error loading reconciliation summaries. Please try again.";
      });
  }

  // Function to display summaries in the table
  function displaySummaries(summaries) {
    summariesTableBody.innerHTML = "";

    summaries.forEach((summary) => {
      const row = document.createElement("tr");

      const startDate = new Date(summary.startDate).toLocaleDateString();
      const endDate = new Date(summary.endDate).toLocaleDateString();

      row.innerHTML = `
        <td class="task-id" title="${summary.taskId}">${summary.taskId}</td>
        <td class="date-format">${startDate} to ${endDate}</td>
        <td>${summary.totalMatched}</td>
        <td class="currency">$${summary.totalDiscrepancy.toFixed(2)}</td>
        <td>${summary.totalUnmatchedInternal}</td>
        <td>${summary.totalUnmatchedBank}</td>
        <td>
          <div class="d-flex">
            <button class="btn btn-sm btn-outline-primary action-btn view-transactions" data-task-id="${summary.taskId}">
              <i class="bi bi-list-ul"></i> Transactions
            </button>
            <button class="btn btn-sm btn-outline-info action-btn view-bank-statements" data-task-id="${summary.taskId}">
              <i class="bi bi-bank"></i> Bank
            </button>
          </div>
        </td>
      `;

      summariesTableBody.appendChild(row);
    });

    // Add event listeners to view buttons
    document.querySelectorAll(".view-transactions").forEach((button) => {
      button.addEventListener("click", function () {
        const taskId = this.getAttribute("data-task-id");
        viewDetails(taskId, "transaction");
      });
    });

    document.querySelectorAll(".view-bank-statements").forEach((button) => {
      button.addEventListener("click", function () {
        const taskId = this.getAttribute("data-task-id");
        viewDetails(taskId, "bank");
      });
    });
  }

  // Function to view transaction or bank statement details
  function viewDetails(taskId, type) {
    // Reset pagination
    detailState.taskId = taskId;
    detailState.type = type;
    detailState.offset = 0;

    // Set modal title based on type
    if (type === "transaction") {
      detailsModalTitle.textContent = `Unmatched Transactions (Task ID: ${taskId})`;
    } else {
      detailsModalTitle.textContent = `Unmatched Bank Statements (Task ID: ${taskId})`;
    }

    // Show the modal
    detailsModal.show();

    // Load the details
    loadDetails();
  }

  // Function to load details based on current detail state
  function loadDetails() {
    // Show loading state
    showElement(detailsLoading);
    hideElement(detailsTable);
    hideElement(noDetails);
    hideElement(detailsPagination);

    let url;
    if (detailState.type === "transaction") {
      url = `/api/reconciliation/summary/${detailState.taskId}/transaction?limit=${detailState.limit}&offset=${detailState.offset}`;
    } else {
      url = `/api/reconciliation/summary/${detailState.taskId}/bank?limit=${detailState.limit}&offset=${detailState.offset}`;
    }

    fetch(url)
      .then((response) => {
        if (!response.ok) {
          throw new Error(`HTTP error! Status: ${response.status}`);
        }
        return response.json();
      })
      .then((data) => {
        const items = data.data;
        detailState.totalCount = data.totalCount;

        // Update table based on type
        if (detailState.type === "transaction") {
          displayTransactions(items);
        } else {
          displayBankStatements(items);
        }

        // Update pagination info
        updatePaginationInfo(
          detailState,
          detailsCurrentRange,
          detailsTotalCount,
          detailsPrevBtn,
          detailsNextBtn,
        );

        // Show appropriate UI elements
        hideElement(detailsLoading);
        if (items.length === 0) {
          showElement(noDetails);
          hideElement(detailsTable);
          hideElement(detailsPagination);
        } else {
          hideElement(noDetails);
          showElement(detailsTable);
          showElement(detailsPagination);
        }
      })
      .catch((error) => {
        console.error(`Error fetching ${detailState.type} details:`, error);
        hideElement(detailsLoading);
        showElement(noDetails);
        noDetails.textContent = `Error loading unmatched ${detailState.type} details. Please try again.`;
      });
  }

  // Function to display transactions in the details table
  function displayTransactions(transactions) {
    // Set up headers
    detailsTableHead.innerHTML = `
      <tr>
        <th>ID</th>
        <th>Amount</th>
        <th>Transaction Time</th>
        <th>Type</th>
        <th>Description</th>
      </tr>
    `;

    // Clear and populate table body
    detailsTableBody.innerHTML = "";
    transactions.forEach((tx) => {
      const row = document.createElement("tr");
      const txTime = new Date(tx.transactionTime).toLocaleString();

      row.innerHTML = `
        <td>${tx.id}</td>
        <td class="currency">$${tx.amount.toFixed(2)}</td>
        <td class="date-format">${txTime}</td>
        <td>${tx.type}</td>
        <td>${tx.description}</td>
      `;

      detailsTableBody.appendChild(row);
    });
  }

  // Function to display bank statements in the details table
  function displayBankStatements(statements) {
    // Set up headers
    detailsTableHead.innerHTML = `
      <tr>
        <th>ID</th>
        <th>Amount</th>
        <th>Date</th>
        <th>Reference</th>
        <th>Bank Name</th>
      </tr>
    `;

    // Clear and populate table body
    detailsTableBody.innerHTML = "";
    statements.forEach((stmt) => {
      const row = document.createElement("tr");
      const stmtDate = new Date(stmt.date).toLocaleDateString();

      row.innerHTML = `
        <td>${stmt.id}</td>
        <td class="currency">$${stmt.amount.toFixed(2)}</td>
        <td class="date-format">${stmtDate}</td>
        <td>${stmt.reference}</td>
        <td>${stmt.bankName}</td>
      `;

      detailsTableBody.appendChild(row);
    });
  }

  // Function to update pagination information and buttons
  function updatePaginationInfo(
    state,
    currentRangeElement,
    totalCountElement,
    prevButton,
    nextButton,
  ) {
    // Calculate current range
    const start = state.offset + 1;
    const end = Math.min(state.offset + state.limit, state.totalCount);

    // Update display
    currentRangeElement.textContent = `${start}-${end}`;
    totalCountElement.textContent = state.totalCount;

    // Update button states
    if (state.offset <= 0) {
      prevButton.parentElement.classList.add("disabled");
    } else {
      prevButton.parentElement.classList.remove("disabled");
    }

    if (state.offset + state.limit >= state.totalCount) {
      nextButton.parentElement.classList.add("disabled");
    } else {
      nextButton.parentElement.classList.remove("disabled");
    }
  }

  // ======== HELPER FUNCTIONS ========

  // UI helper functions
  function showElement(element) {
    element.classList.remove("d-none");
  }

  function hideElement(element) {
    element.classList.add("d-none");
  }

  function updateProgress(progressElement, percent) {
    const bar = progressElement.querySelector(".progress-bar");
    bar.style.width = percent + "%";
    bar.setAttribute("aria-valuenow", percent);
  }

  function showError(element, message) {
    element.textContent = message;
    showElement(element);
  }

  function showModal(title, message) {
    modalTitle.textContent = title;
    modalBody.textContent = message;
    notificationModal.show();
  }
});
