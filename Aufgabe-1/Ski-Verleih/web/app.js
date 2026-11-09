const state = {
  skis: [],
  customers: [],
  rentals: [],
  dailyRateCents: 2000,
  editingSkiID: null,
  editingCustomerID: null,
};

const elements = {
  notice: document.querySelector("#notice"),
  skiForm: document.querySelector("#ski-form"),
  skiType: document.querySelector("#ski-type"),
  skiCancel: document.querySelector("#ski-cancel"),
  skiRows: document.querySelector("#ski-rows"),
  skiEmpty: document.querySelector("#ski-empty"),
  customerForm: document.querySelector("#customer-form"),
  customerName: document.querySelector("#customer-name"),
  documentType: document.querySelector("#document-type"),
  documentNumber: document.querySelector("#document-number"),
  customerSubmit: document.querySelector("#customer-submit"),
  customerCancel: document.querySelector("#customer-cancel"),
  customerRows: document.querySelector("#customer-rows"),
  customerEmpty: document.querySelector("#customer-empty"),
  rentalForm: document.querySelector("#rental-form"),
  rentalSki: document.querySelector("#rental-ski"),
  rentalCustomer: document.querySelector("#rental-customer"),
  rentalStartDate: document.querySelector("#rental-start-date"),
  rentalDays: document.querySelector("#rental-days"),
  rentalPrice: document.querySelector("#rental-price"),
  rentalSubmit: document.querySelector("#rental-submit"),
  rentalRows: document.querySelector("#rental-rows"),
  rentalEmpty: document.querySelector("#rental-empty"),
  dailyRateForm: document.querySelector("#daily-rate-form"),
  dailyRate: document.querySelector("#daily-rate"),
};

function node(tag, className, text) {
  const element = document.createElement(tag);
  if (className) element.className = className;
  if (text !== undefined) element.textContent = text;
  return element;
}

function addDays(date, days) {
  if (!date || !Number.isSafeInteger(days) || days < 1) return "";
  const [year, month, day] = date.split("-").map(Number);
  const result = new Date(year, month - 1, day + days);
  return `${result.getFullYear()}-${`${result.getMonth() + 1}`.padStart(2, "0")}-${`${result.getDate()}`.padStart(2, "0")}`;
}

async function request(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: {
      ...(options.body ? { "Content-Type": "application/json" } : {}),
      ...options.headers,
    },
  });
  const result = await response.json();
  if (!response.ok) throw new Error(result.error || "Die Anfrage ist fehlgeschlagen.");
  return result;
}

function showNotice(message, isError = false) {
  elements.notice.textContent = message;
  elements.notice.classList.toggle("error", isError);
  elements.notice.hidden = false;
  window.clearTimeout(showNotice.timeout);
  showNotice.timeout = window.setTimeout(() => { elements.notice.hidden = true; }, 5000);
}

function addCell(row, value) {
  row.append(node("td", "", value));
}

function makeAction(label, action, className = "button button-small") {
  const button = node("button", className, label);
  button.type = "button";
  button.addEventListener("click", action);
  return button;
}

function renderDashboard(dashboard) {
  document.querySelector("#available-count").textContent = dashboard.availableSkis;
  document.querySelector("#rented-count").textContent = dashboard.rentedSkis;
  document.querySelector("#total-count").textContent = dashboard.totalSkis;
  document.querySelector("#customer-count").textContent = dashboard.customers;
  document.querySelector("#ski-total").textContent = `${dashboard.totalSkis} Ski`;
  document.querySelector("#customer-total").textContent = `${dashboard.customers} Kunden`;
}

function renderSkis() {
  elements.skiRows.replaceChildren();
  elements.skiEmpty.hidden = state.skis.length > 0;
  for (const ski of state.skis) {
    const row = document.createElement("tr");
    addCell(row, `#${ski.id}`);
    addCell(row, ski.type);
    const status = node("span", `badge ${ski.rented ? "badge-rented" : "badge-available"}`, ski.rented ? "Ausgeliehen" : "Verfügbar");
    const statusCell = document.createElement("td");
    statusCell.append(status);
    row.append(statusCell);

    const actions = document.createElement("td");
    actions.append(makeAction("Bearbeiten", () => beginSkiEdit(ski)));
    row.append(actions);
    elements.skiRows.append(row);
  }

  renderAvailableSkis();
}

function renderAvailableSkis() {
  const startDate = elements.rentalStartDate.value;
  const days = Number(elements.rentalDays.value);
  const dueDate = addDays(startDate, days);
  const selectedSkiID = elements.rentalSki.value;
  const validPeriod = startDate !== "" && Number.isSafeInteger(days) && days > 0;
  const availableSkis = validPeriod
    ? state.skis.filter((ski) => !state.rentals.some((rental) => (
      rental.skiId === ski.id
      && !rental.returnDate
      && startDate < datePart(rental.dueDate)
      && datePart(rental.startDate) < dueDate
    )))
    : [];
  replaceOptions(
    elements.rentalSki,
    availableSkis.map((ski) => [`${ski.id}`, `#${ski.id} · ${ski.type}`]),
    validPeriod ? "Kein Ski im Zeitraum verfügbar" : "Bitte Zeitraum eingeben",
    selectedSkiID,
    true,
  );
}

function renderCustomers() {
  elements.customerRows.replaceChildren();
  elements.customerEmpty.hidden = state.customers.length > 0;
  for (const customer of state.customers) {
    const row = document.createElement("tr");
    addCell(row, `#${customer.id}`);
    addCell(row, customer.name);
    addCell(row, `${customer.documentType} · ${customer.documentNumber}`);
    const actions = document.createElement("td");
    actions.append(makeAction("Bearbeiten", () => beginCustomerEdit(customer)));
    row.append(actions);
    elements.customerRows.append(row);
  }
  replaceOptions(elements.rentalCustomer, state.customers.map((customer) => [`${customer.id}`, `#${customer.id} · ${customer.name}`]), "Kein Kunde registriert");
}

function renderRentals() {
  const skis = new Map(state.skis.map((ski) => [ski.id, ski]));
  const customers = new Map(state.customers.map((customer) => [customer.id, customer]));
  elements.rentalRows.replaceChildren();
  elements.rentalEmpty.hidden = state.rentals.length > 0;
  for (const rental of [...state.rentals].reverse()) {
    const row = document.createElement("tr");
    addCell(row, `#${rental.skiId} · ${skis.get(rental.skiId)?.type || "Ski"}`);
    addCell(row, customers.get(rental.customerId)?.name || `Kunde #${rental.customerId}`);
    const startDate = datePart(rental.startDate);
    addCell(row, formatDate(startDate));
    addCell(row, `${formatDate(rental.dueDate)} (${rental.days} ${rental.days === 1 ? "Tag" : "Tage"})`);
    const priceCell = document.createElement("td");
    priceCell.textContent = formatMoney(rental.totalPriceCents);
    priceCell.title = `${formatMoney(rental.dailyRateCents)} pro Tag`;
    row.append(priceCell);
    const statusCell = document.createElement("td");
    const returned = Boolean(rental.returnDate);
    const reserved = !returned && startDate > localDateValue();
    const statusText = returned ? `Zurückgegeben · ${formatDate(rental.returnDate)}` : reserved ? "Reserviert" : "Ausgeliehen";
    const statusClass = returned ? "badge-returned" : reserved ? "badge-reserved" : "badge-rented";
    statusCell.append(node("span", `badge ${statusClass}`, statusText));
    row.append(statusCell);
    const actions = document.createElement("td");
    if (!returned && !reserved) {
      actions.append(makeAction("Rückgabe", () => returnSki(rental.skiId), "button button-small button-return"));
    } else {
      actions.textContent = "—";
    }
    row.append(actions);
    elements.rentalRows.append(row);
  }
}

function replaceOptions(select, options, emptyText, preferredValue = "", enablePlaceholder = false) {
  select.replaceChildren();
  const placeholder = node("option", "", emptyText);
  placeholder.value = "";
  placeholder.disabled = !enablePlaceholder;
  placeholder.selected = true;
  select.append(placeholder);
  for (const [value, label] of options) {
    const option = node("option", "", label);
    option.value = value;
    select.append(option);
  }
  select.disabled = options.length === 0;
  if (options.some(([value]) => value === preferredValue)) {
    select.value = preferredValue;
  }
}

function formatDate(value) {
  const [year, month, day] = datePart(value).split("-").map(Number);
  return new Intl.DateTimeFormat("de-AT", { dateStyle: "medium" }).format(new Date(year, month - 1, day));
}

function datePart(value) {
  return value.slice(0, 10);
}

function localDateValue(date = new Date()) {
  const year = date.getFullYear();
  const month = `${date.getMonth() + 1}`.padStart(2, "0");
  const day = `${date.getDate()}`.padStart(2, "0");
  return `${year}-${month}-${day}`;
}

function formatMoney(cents) {
  return new Intl.NumberFormat("de-AT", { style: "currency", currency: "EUR" }).format(cents / 100);
}

function renderPricePreview() {
  const days = Number(elements.rentalDays.value);
  const rate = state.dailyRateCents;
  elements.rentalPrice.textContent = formatMoney(Number.isSafeInteger(days) && days > 0 && Number.isSafeInteger(rate) && rate >= 0 ? days * rate : 0);
  const isReservation = elements.rentalStartDate.value > localDateValue();
  elements.rentalSubmit.textContent = isReservation ? "Reservierung erstellen" : "Ausleihe erfassen";
}

async function refresh() {
  const [dashboard, skis, customers, rentals, config] = await Promise.all([
    request("/api/dashboard"),
    request("/api/skis"),
    request("/api/customers"),
    request("/api/rentals"),
    request("/api/config"),
  ]);
  state.skis = skis;
  state.customers = customers;
  state.rentals = rentals;
  state.dailyRateCents = config.dailyRateCents;
  elements.dailyRate.value = (state.dailyRateCents / 100).toFixed(2);
  renderDashboard(dashboard);
  renderSkis();
  renderCustomers();
  renderRentals();
  renderPricePreview();
}

function beginSkiEdit(ski) {
  state.editingSkiID = ski.id;
  elements.skiType.value = ski.type;
  elements.skiForm.querySelector("button[type=submit]").textContent = "Änderung speichern";
  elements.skiCancel.hidden = false;
  elements.skiType.focus();
}

function resetSkiForm() {
  state.editingSkiID = null;
  elements.skiForm.reset();
  elements.skiForm.querySelector("button[type=submit]").textContent = "＋ Ski hinzufügen";
  elements.skiCancel.hidden = true;
}

function beginCustomerEdit(customer) {
  state.editingCustomerID = customer.id;
  elements.customerName.value = customer.name;
  elements.documentType.value = customer.documentType;
  elements.documentNumber.value = customer.documentNumber;
  elements.customerSubmit.textContent = "Änderung speichern";
  elements.customerCancel.hidden = false;
  elements.customerName.focus();
}

function resetCustomerForm() {
  state.editingCustomerID = null;
  elements.customerForm.reset();
  elements.customerSubmit.textContent = "Kunde registrieren";
  elements.customerCancel.hidden = true;
}

async function returnSki(skiId) {
  try {
    await request("/api/rentals/return", { method: "POST", body: JSON.stringify({ skiId }) });
    await refresh();
    showNotice(`Ski #${skiId} wurde als zurückgegeben erfasst.`);
  } catch (error) {
    showNotice(error.message, true);
  }
}

elements.skiForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  try {
    const payload = { type: elements.skiType.value };
    if (state.editingSkiID === null) {
      await request("/api/skis", { method: "POST", body: JSON.stringify(payload) });
      showNotice("Ski wurde zum Bestand hinzugefügt.");
    } else {
      await request(`/api/skis/${state.editingSkiID}`, { method: "PUT", body: JSON.stringify(payload) });
      showNotice(`Ski #${state.editingSkiID} wurde aktualisiert.`);
    }
    resetSkiForm();
    await refresh();
  } catch (error) {
    showNotice(error.message, true);
  }
});

elements.skiCancel.addEventListener("click", resetSkiForm);

elements.customerForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = {
    name: elements.customerName.value,
    documentType: elements.documentType.value,
    documentNumber: elements.documentNumber.value,
  };
  try {
    if (state.editingCustomerID === null) {
      await request("/api/customers", { method: "POST", body: JSON.stringify(payload) });
      showNotice("Kunde wurde registriert.");
    } else {
      await request(`/api/customers/${state.editingCustomerID}`, { method: "PUT", body: JSON.stringify(payload) });
      showNotice(`Kundendaten für #${state.editingCustomerID} wurden aktualisiert.`);
    }
    resetCustomerForm();
    await refresh();
  } catch (error) {
    showNotice(error.message, true);
  }
});

elements.customerCancel.addEventListener("click", resetCustomerForm);

elements.rentalForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const startDate = elements.rentalStartDate.value;
  const payload = {
    skiId: Number(elements.rentalSki.value),
    customerId: Number(elements.rentalCustomer.value),
    days: Number(elements.rentalDays.value),
  };
  try {
    const isReservation = startDate > localDateValue();
    const rental = isReservation
      ? await request("/api/reservations", { method: "POST", body: JSON.stringify({ ...payload, startDate }) })
      : await request("/api/rentals", { method: "POST", body: JSON.stringify(payload) });
    elements.rentalForm.reset();
    setDefaultRentalDate();
    elements.rentalDays.value = "1";
    await refresh();
    showNotice(isReservation
      ? `Ski #${rental.skiId} wurde vom ${formatDate(rental.startDate)} bis ${formatDate(rental.dueDate)} reserviert. Gesamtpreis: ${formatMoney(rental.totalPriceCents)}.`
      : `Ski #${rental.skiId} wurde bis ${formatDate(rental.dueDate)} ausgeliehen. Gesamtpreis: ${formatMoney(rental.totalPriceCents)}.`);
  } catch (error) {
    showNotice(error.message, true);
  }
});

elements.dailyRateForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const amount = Number(elements.dailyRate.value);
  if (!Number.isFinite(amount) || amount < 0 || Math.round(amount * 100) !== amount * 100) {
    showNotice("Bitte gib einen gültigen Tarif mit höchstens zwei Nachkommastellen ein.", true);
    return;
  }
  try {
    const config = await request("/api/config", {
      method: "PUT",
      body: JSON.stringify({ dailyRateCents: Math.round(amount * 100) }),
    });
    state.dailyRateCents = config.dailyRateCents;
    renderPricePreview();
    showNotice(`Der Tagestarif beträgt jetzt ${formatMoney(config.dailyRateCents)}.`);
  } catch (error) {
    showNotice(error.message, true);
  }
});

elements.rentalDays.addEventListener("input", renderPricePreview);
elements.rentalDays.addEventListener("input", renderAvailableSkis);
elements.rentalStartDate.addEventListener("change", () => {
  renderPricePreview();
  renderAvailableSkis();
});

function setDefaultRentalDate() {
  const today = localDateValue();
  elements.rentalStartDate.min = today;
  elements.rentalStartDate.value = today;
}

document.querySelector("#today").textContent = new Intl.DateTimeFormat("de-AT", { dateStyle: "full" }).format(new Date());
setDefaultRentalDate();
refresh().catch((error) => showNotice(`Daten konnten nicht geladen werden: ${error.message}`, true));
