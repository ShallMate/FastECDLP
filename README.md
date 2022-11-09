# This repository is FastDec combined with Exp-ElGamal additively homomorphic encryption scheme.
## Points to note before using this code are as follows (we believe its efficiency will surprise you):
  
(1) We did not upload the precomputed table for running the BSGS because it is too large, you can generate one from the code in the genlist folder (`go run genlist.go > Tx28.txt`). At the same time you need to modify the path to this file in the init function in ciphering.go.  

(2) You can get a feel for the efficiency of the improved Exp-ElGamal by running test.go in the test folder. Since loading the precomputed table in step (2) takes some time (maybe a few minutes), you can set Jlen to be less than 24 so you don't wait too long.

(3) If you want to change the length of the plaintext (of course, this is the work after the code runs successfully), you can do it by changing the two variables Ilen and Jlen in ciphering.go, Ilen+Jlen is equal to the length of the plaintext, and our improved-Exp-ElGamal supports negative arithmetic.  

If you have any questions, please contact: s200201071@stu.cqupt.edu.cn


