// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/4/Mult.asm

// Multiplies R0 and R1 and stores the result in R2.
// (R0, R1, R2 refer to RAM[0], RAM[1], and RAM[2], respectively.)
// The algorithm is based on repetitive addition.

@i
M=1 // i = 1

@R2
M=0 // set R2 result to initial value of 0

(LOOP)
// check exit condition
@i
D=M // D = whatever i is this iteration
@R1 // get the value of R1, only add R0 into R2 this R1 times
D=D-M // D - R1 value
@END
D;JGT // if D - R1 > 0, go to end

// sum
@R0
D=M // R0 is added R1 times and stored in R2
@R2
M=D+M // R2 = R2 + R0
@i
M=M+1 // increment i for next iteration

@LOOP
0;JMP

(END)
@END
0;JMP