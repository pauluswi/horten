

Single Machine Locks Output:

Processed transaction of -200.00 on account 11111. New balance: 800.00
Processed transaction of 100.00 on account 11111. New balance: 900.00
Processed transaction of 300.00 on account 11111. New balance: 1200.00
Processed transaction of -500.00 on account 22222. New balance: 1500.00
Insufficient funds for account 22222

Final Account Balances:
Account 11111: 1200.00
Account 22222: 1500.00



Distributed Locks Output:

Processed transaction of 100.00 on account 11111. New balance: 1100.00
Insufficient funds for account 22222. Transaction skipped.
Processed transaction of 300.00 on account 11111. New balance: 1400.00
Processed transaction of -500.00 on account 22222. New balance: 1500.00
Processed transaction of -200.00 on account 11111. New balance: 1200.00

Final Account Balances:
Account 11111: 1200.00
Account 22222: 1500.00


