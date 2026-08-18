# FinanceOS

FinanceOS is a personal-finance tracking and decision-support product. Its language distinguishes observed financial facts from classifications, forecasts, and optional explanations.

## Language

**User**:
The person who owns a private FinanceOS workspace and every financial record within it.
_Avoid_: Customer, client

**Connection**:
A user's authorization for FinanceOS to import financial data through Pluggy/Open Finance. One connection may expose multiple accounts.
_Avoid_: Account, bank

**Account**:
A source of balances and transactions, either imported through a connection or maintained manually.
_Avoid_: Connection, investment, portfolio

**Transaction**:
An imported or manually recorded financial movement associated with an account.
_Avoid_: Payment, transfer

**Category**:
The financial-purpose label assigned to a transaction.
_Avoid_: Tag, merchant

**Rule**:
A reusable classification association learned for one user or provided as a global default.
_Avoid_: Category, alert rule

**Commitment**:
A known future obligation, recurring or finite, such as a subscription, installment, or debt.
_Avoid_: Transaction, goal

**Alert**:
An actionable condition detected from the user's financial data and linked to the facts that caused it.
_Avoid_: Analysis, notification

**Goal**:
A measurable financial objective whose progress combines observed data with an explicit history of manual adjustments.
_Avoid_: Simulation, commitment

**Simulation**:
A counterfactual projection showing how a proposed decision could affect the user's finances; it is never an executed transaction.
_Avoid_: Forecast, investment recommendation

**Analysis**:
A deterministic interpretation of financial data for a stated period, with disclosed data quality and confidence.
_Avoid_: Alert, report entry

**Pierre**:
The optional conversational explainer for open questions and already-calculated FinanceOS results; Pierre is not the source of financial calculations.
_Avoid_: Financial adviser, calculation engine
