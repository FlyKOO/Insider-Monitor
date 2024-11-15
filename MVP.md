# MVP Project Plan - Solana Wallet Monitoring Tool

## Simplified Goal
Create a basic Go tool that monitors specific Solana wallet addresses for significant new token holdings.
The MVP will focus on essential functionality to quickly identify and alert on new token additions using a lightweight and simple approach.

---

## Core Features for MVP

1. **Basic Wallet Monitoring**
   - Monitor multiple wallet addresses.
   - Fetch all token accounts and balances for specified wallets.

2. **New Token Detection**
   - Store token accounts from the initial scan in memory and a JSON file.
   - Compare the latest token accounts with previously stored data to identify new additions.

3. **Basic Alerting**
   - Log new token additions to the console or a simple log file.
        - Create a placeholder alerting system for new token holdings.

4. **Persistent Storage**
   - Use a JSON file to store token account data.

---

## Simplified Project Plan

### Phase 1: Initial Setup
- **Goal**: Quickly set up the project structure and install dependencies.
- **Tasks**:
  - Initialize the Go project and install the necessary SDK.
  - Write basic code to connect to the Solana blockchain and test fetching token accounts.

**Estimated Time**: 1 day

### Phase 2: Token Account Fetching
- **Goal**: Implement functionality to fetch token accounts and their balances for each wallet.
- **Tasks**:
  - Develop a simple function to retrieve token accounts for a given wallet address.
  - Iterate over multiple wallet addresses to fetch and display token balances.

**Estimated Time**: 1 to 2 days

### Phase 3: Data Persistence
- **Goal**: Store token account data in a JSON file.
- **Tasks**:
  - Implement functions to save and load token account data.
  - Ensure the JSON file is updated with new token accounts when detected.

**Estimated Time**: 1 day

### Phase 4: New Token Detection
- **Goal**: Identify new token holdings and alert on changes.
- **Tasks**:
  - Compare current token accounts with stored data.
  - Log new token additions to the console or a log file.

**Estimated Time**: 1 day

### Phase 5: Testing and Finalization
- **Goal**: Test the tool with real wallet addresses and ensure it works as expected.
- **Tasks**:
  - Run tests to verify functionality.
  - Fix any issues or bugs that arise during testing.

**Estimated Time**: 1 to 2 days

---

## Total Estimated Time for MVP
**Approximately 4 to 7 days**

---

## Future Enhancements (Post-MVP)
1. **Transaction History Analysis**: Implement the feature to track token account creation times to avoid redundant alerts.
2. **Advanced Alerting**: Add email or push notifications for alerts.
3. **Performance Improvements**: Optimize data fetching and implement better error handling.
4. **Database Migration**: Upgrade to a more robust database if needed.

---

## Conclusion
This MVP will allow you to quickly deploy a functional tool that can monitor Solana wallet addresses for new token holdings.
By focusing on essential features and deferring complex components,
you can get your product out quickly and iteratively improve it based on feedback and additional requirements.
