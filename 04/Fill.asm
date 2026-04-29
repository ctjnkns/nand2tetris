// This file is part of www.nand2tetris.org
// and the book "The Elements of Computing Systems"
// by Nisan and Schocken, MIT Press.
// File name: projects/4/Fill.asm

// Runs an infinite loop that listens to the keyboard input. 
// When a key is pressed (any key), the program blackens the screen,
// i.e. writes "black" in every pixel. When no key is pressed, 
// the screen should be cleared.

// add the upper bound as a constant to check later
@8192
D=A
@n
M=D

(RESETI)
@i
M=0
@LOOP
0;JMP

(FILLBLACK)
@SCREEN
D=A
@i
D=D+M
A=D
// fill black
M=-1
// increment i
@i
M=M+1
@LOOP
0;JMP

(FILLWHITE)
@SCREEN
D=A
@i
D=D+M
A=D
// fill white
M=0
// increment i
@i
M=M+1
@LOOP
0;JMP

(LOOP)
// check if i is out of bounds
@i
D=M
@n
D=D-M
@RESETI
D;JGE

// get keyboard input
@KBD
D=M

@FILLBLACK
D;JGT

@FILLWHITE
D;JEQ

// always loop
@LOOP
0;JMP