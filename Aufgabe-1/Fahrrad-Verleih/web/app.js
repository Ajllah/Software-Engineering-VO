const state = {
  bikes: [],
  customers: [],
  bookings: [],
  hourlyRateCents: 500,
  editingBikeID: null,
  editingCustomerID: null,
};

const $ = (selector) => document.querySelector(selector);
const elements = {
  notice: $("#notice"),
  bikeForm: $("#bike-form"),
  bikeType: $("#bike-type"),
  bikeSubmit: $("#bike-submit"),
  bikeCancel: $("#bike-cancel"),
  bikeRows: $("#bike-rows"),
  bikeEmpty: $("#bike-empty"),
  customerForm: $("#customer-form"),
  customerName: $("#customer-name"),
  documentType: $("#document-type"),
  documentNumber: $("#document-number"),
  customerSubmit: $("#customer-submit"),
  customerCancel: $("#customer-cancel"),
  customerRows: $("#customer-rows"),
  customerEmpty: $("#customer-empty"),
  bookingForm: $("#booking-form"),
  bookingBike: $("#booking-bike"),
  bookingCustomer: $("#booking-customer"),
  bookingStart: $("#booking-start"),
  bookingEnd: $("#booking-end"),
  bookingPrice: $("#booking-price"),
  priceHours: $("#price-hours"),
  bookingRows: $("#booking-rows"),
  bookingEmpty: $("#booking-empty"),
  rateForm: $("#rate-form"),
  hourlyRate: $("#hourly-rate"),
};

function node(tag, className, text) {
  const element = document.createElement(tag);
  if (className) element.className = className;
  if (text !== undefined) element.textContent = text;
  return element;
}

async function request(path, options = {}) {
  const response = await fetch(path, {
    ...options,
    headers: { ...(options.body ? { "Content-Type": "application/json" } : {}), ...options.headers },
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

function formatMoney(cents) {
  return new Intl.NumberFormat("de-AT", { style: "currency", currency: "EUR" }).format(cents / 100);
}

function formatDateTime(value) {
  return new Intl.DateTimeFormat("de-AT", { dateStyle: "short", timeStyle: "short" }).format(new Date(value));
}

function localDateTimeInput(date) {
  const local = new Date(date.getTime() - date.getTimezoneOffset() * 60000);
  return local.toISOString().slice(0, 16);
}

function parseLocalDateTime(value) {
  if (!value) return null;
  return new Date(`${value}:00`);
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

function replaceOptions(select, options, emptyText, preferredValue = "") {
  select.replaceChildren();
  const placeholder = node("option", "", emptyText);
  placeholder.value = "";
  placeholder.disabled = true;
  placeholder.selected = true;
  select.append(placeholder);
  for (const [value, label] of options) {
    const option = node("option", "", label);
    option.value = value;
    select.append(option);
  }
  if (options.some(([value]) => `${value}` === `${preferredValue}`)) select.value = `${preferredValue}`;
  select.disabled = options.length === 0;
}

function renderDashboard(dashboard) {
  $("#available-count").textContent = dashboard.availableBikes;
  $("#rented-count").textContent = dashboard.rentedBikes;
  $("#total-count").textContent = dashboard.totalBikes;
  $("#customer-count").textContent = dashboard.customers;
  $("#bike-total").textContent = `${dashboard.totalBikes} Fahrräder`;
  $("#customer-total").textContent = `${dashboard.customers} Kunden`;
  $("#booking-total").textContent = `${state.bookings.length} Buchungen`;
}

function effectiveEnd(booking) {
  const planned = new Date(booking.endTime);
  const returned = booking.returnTime ? new Date(booking.returnTime) : null;
  return returned && returned < planned ? returned : planned;
}

function renderBikes() {
  elements.bikeRows.replaceChildren();
  elements.bikeEmpty.hidden = state.bikes.length > 0;
  for (const bike of state.bikes) {
    const row = document.createElement("tr");
    addCell(row, `#${bike.id}`);
    addCell(row, bike.type);
    const actions = document.createElement("td");
    actions.append(makeAction("Bearbeiten", () => beginBikeEdit(bike)));
    row.append(actions);
    elements.bikeRows.append(row);
  }
  renderAvailableBikes();
}

function renderAvailableBikes(startOverride = null) {
  const start = startOverride || parseLocalDateTime(elements.bookingStart.value);
  const end = parseLocalDateTime(elements.bookingEnd.value);
  const previous = elements.bookingBike.value;
  const validPeriod = start && end && end > start;
  const available = validPeriod ? state.bikes.filter((bike) => !state.bookings.some((booking) => (
    booking.bikeId === bike.id
    && new Date(booking.startTime) < end
    && start < effectiveEnd(booking)
  ))) : [];
  replaceOptions(
    elements.bookingBike,
    available.map((bike) => [bike.id, `#${bike.id} · ${bike.type}`]),
    validPeriod ? "Kein Fahrrad im Zeitraum verfügbar" : "Bitte Zeitraum eingeben",
    previous,
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
  replaceOptions(
    elements.bookingCustomer,
    state.customers.map((customer) => [customer.id, `#${customer.id} · ${customer.name}`]),
    "Kein Kunde registriert",
    elements.bookingCustomer.value,
  );
}

function renderBookings() {
  const bikes = new Map(state.bikes.map((bike) => [bike.id, bike]));
  const customers = new Map(state.customers.map((customer) => [customer.id, customer]));
  elements.bookingRows.replaceChildren();
  elements.bookingEmpty.hidden = state.bookings.length > 0;
  for (const booking of [...state.bookings].reverse()) {
    const row = document.createElement("tr");
    addCell(row, `#${booking.bikeId} · ${bikes.get(booking.bikeId)?.type || "Fahrrad"}`);
    addCell(row, customers.get(booking.customerId)?.name || `Kunde #${booking.customerId}`);
    addCell(row, `${formatDateTime(booking.startTime)} – ${formatDateTime(booking.endTime)}`);
    const durationMs = new Date(booking.endTime) - new Date(booking.startTime);
    const minutes = Math.round(durationMs / 60000);
    addCell(row, `${Math.floor(minutes / 60)} h ${minutes % 60} min`);
    const price = document.createElement("td");
    price.textContent = formatMoney(booking.totalPriceCents);
    price.title = `${formatMoney(booking.hourlyRateCents)} pro Stunde`;
    row.append(price);

    const returned = Boolean(booking.returnTime);
    const future = new Date(booking.startTime) > new Date();
    const status = returned ? "Zurückgegeben" : future ? "Reserviert" : new Date(booking.endTime) > new Date() ? "Ausgeliehen" : "Rückgabe offen";
    const badge = node("span", `badge ${returned ? "badge-returned" : future ? "badge-reserved" : "badge-rented"}`, status);
    const statusCell = document.createElement("td");
    statusCell.append(badge);
    row.append(statusCell);

    const actions = document.createElement("td");
    if (!returned && !future) {
      actions.append(makeAction("Rückgabe", () => returnBooking(booking.id), "button button-small button-return"));
    } else {
      actions.textContent = "—";
    }
    row.append(actions);
    elements.bookingRows.append(row);
  }
  $("#booking-total").textContent = `${state.bookings.length} Buchungen`;
}

function renderPrice() {
  const start = parseLocalDateTime(elements.bookingStart.value);
  const end = parseLocalDateTime(elements.bookingEnd.value);
  if (!start || !end || end <= start) {
    elements.priceHours.textContent = "Voraussichtlicher Gesamtpreis";
    elements.bookingPrice.textContent = formatMoney(0);
    return;
  }
  const minutes = Math.ceil((end - start) / 60000);
  const hours = Math.ceil(minutes / 60);
  elements.priceHours.textContent = `${hours} ${hours === 1 ? "Stunde" : "Stunden"} · voraussichtlicher Gesamtpreis`;
  elements.bookingPrice.textContent = formatMoney(hours * state.hourlyRateCents);
}

function render() {
  renderDashboard({ ...state.dashboard, totalBikes: state.bikes.length, customers: state.customers.length });
  renderBikes();
  renderCustomers();
  renderBookings();
  renderPrice();
}

async function refresh() {
  const [dashboard, bikes, customers, bookings, config] = await Promise.all([
    request("/api/dashboard"), request("/api/bikes"), request("/api/customers"),
    request("/api/bookings"), request("/api/config"),
  ]);
  state.dashboard = dashboard;
  state.bikes = bikes;
  state.customers = customers;
  state.bookings = bookings;
  state.hourlyRateCents = config.hourlyRateCents;
  elements.hourlyRate.value = (state.hourlyRateCents / 100).toFixed(2);
  render();
}

function resetBikeForm() {
  state.editingBikeID = null;
  elements.bikeForm.reset();
  elements.bikeSubmit.textContent = "+ Fahrrad hinzufügen";
  elements.bikeCancel.hidden = true;
}

function beginBikeEdit(bike) {
  state.editingBikeID = bike.id;
  elements.bikeType.value = bike.type;
  elements.bikeSubmit.textContent = "Änderung speichern";
  elements.bikeCancel.hidden = false;
  elements.bikeType.focus();
}

function resetCustomerForm() {
  state.editingCustomerID = null;
  elements.customerForm.reset();
  elements.customerSubmit.textContent = "Kunde registrieren";
  elements.customerCancel.hidden = true;
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

elements.bikeForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  try {
    const editing = state.editingBikeID !== null;
    await request(editing ? `/api/bikes/${state.editingBikeID}` : "/api/bikes", {
      method: editing ? "PUT" : "POST", body: JSON.stringify({ type: elements.bikeType.value }),
    });
    resetBikeForm();
    await refresh();
    showNotice(editing ? "Fahrrad aktualisiert." : "Fahrrad hinzugefügt.");
  } catch (error) { showNotice(error.message, true); }
});

elements.bikeCancel.addEventListener("click", resetBikeForm);

elements.customerForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const payload = {
    name: elements.customerName.value,
    documentType: elements.documentType.value,
    documentNumber: elements.documentNumber.value,
  };
  try {
    const editing = state.editingCustomerID !== null;
    await request(editing ? `/api/customers/${state.editingCustomerID}` : "/api/customers", {
      method: editing ? "PUT" : "POST", body: JSON.stringify(payload),
    });
    resetCustomerForm();
    await refresh();
    showNotice(editing ? "Kundendaten aktualisiert." : "Kunde registriert.");
  } catch (error) { showNotice(error.message, true); }
});

elements.customerCancel.addEventListener("click", resetCustomerForm);

elements.bookingForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const startNow = event.submitter?.id === "booking-now";
  if (startNow) {
    renderAvailableBikes(new Date());
    if (!elements.bookingBike.value) {
      showNotice("Für den aktuellen Zeitpunkt ist kein Fahrrad verfügbar.", true);
      return;
    }
    if (parseLocalDateTime(elements.bookingEnd.value) <= new Date()) {
      showNotice("Das Buchungsende muss in der Zukunft liegen.", true);
      return;
    }
  }
  try {
    const booking = await request("/api/bookings", {
      method: "POST",
      body: JSON.stringify({
        bikeId: Number(elements.bookingBike.value),
        customerId: Number(elements.bookingCustomer.value),
        startNow,
        startTime: elements.bookingStart.value,
        endTime: elements.bookingEnd.value,
      }),
    });
    await refresh();
    showNotice(`${startNow ? "Ausleihe" : "Reservierung"} #${booking.id} erfasst · ${formatMoney(booking.totalPriceCents)}.`);
  } catch (error) { showNotice(error.message, true); }
});

async function returnBooking(id) {
  try {
    await request(`/api/bookings/${id}/return`, { method: "POST" });
    await refresh();
    showNotice("Rückgabe erfasst. Das Fahrrad ist ab jetzt wieder verfügbar.");
  } catch (error) { showNotice(error.message, true); }
}

elements.rateForm.addEventListener("submit", async (event) => {
  event.preventDefault();
  const rate = Number(elements.hourlyRate.value);
  if (!Number.isFinite(rate) || rate < 0 || !Number.isInteger(rate * 100)) {
    showNotice("Bitte einen gültigen Tarif mit höchstens zwei Nachkommastellen eingeben.", true);
    return;
  }
  try {
    await request("/api/config", { method: "PUT", body: JSON.stringify({ hourlyRateCents: Math.round(rate * 100) }) });
    await refresh();
    showNotice("Stundentarif gespeichert. Bestehende Buchungen bleiben unverändert.");
  } catch (error) { showNotice(error.message, true); }
});

for (const input of [elements.bookingStart, elements.bookingEnd]) {
  input.addEventListener("input", () => { renderAvailableBikes(); renderPrice(); });
  input.addEventListener("change", () => { renderAvailableBikes(); renderPrice(); });
}

const startDefault = new Date(Date.now() + 5 * 60000);
startDefault.setSeconds(0, 0);
const endDefault = new Date(startDefault.getTime() + 60 * 60000);
elements.bookingStart.value = localDateTimeInput(startDefault);
elements.bookingEnd.value = localDateTimeInput(endDefault);
$("#today").textContent = new Intl.DateTimeFormat("de-AT", { dateStyle: "full" }).format(new Date());
refresh().catch((error) => showNotice(`Anwendung konnte nicht geladen werden: ${error.message}`, true));
