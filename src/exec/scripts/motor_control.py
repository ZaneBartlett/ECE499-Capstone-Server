# Simple demo of of the PCA9685 PWM servo/LED controller library.
# This will move channel 0 from min to max position repeatedly.
# Author: Tony DiCola
# License: Public Domain
from __future__ import division
import time
import sys

# Import the PCA9685 module.
import Adafruit_PCA9685

# Initialise the PCA9685 using the default address (0x40).
pwm = Adafruit_PCA9685.PCA9685()

# Configure min and max servo pulse lengths
servo_min = 150  # Min pulse length out of 4096
servo_max = 600  # Max pulse length out of 4096

# Set frequency to 60hz, good for servos.
pwm.set_pwm_freq(60)

print('Moving servo on channel 0, press Ctrl-C to quit...')
# while True:
 if sys.argv[1] == "0":
    loopCount = sys.argv[2]
    while loopCount is not 0:
        # Move servo on channel 0 between extremes.
        pwm.set_pwm(0, 0, servo_min)
        time.sleep(1)
        pwm.set_pwm(0, 0, servo_max)
        time.sleep(2)
        print("Servo 0 Completed! ")  
        loopCount -= 1

elif sys.argv[1] == "1":
    loopCount = sys.argv[2]
    while loopCount is not 0:
        # Move servo on channel 1 between extremes.
        pwm.set_pwm(1, 0, servo_min)
        time.sleep(1)
        pwm.set_pwm(0, 0, servo_max)
        time.sleep(2)
        print("Servo 1 Completed! ")
        loopCount -= 1

elif sys.argv[1] == "2":
    loopCount = sys.argv[2]
    while loopCount is not 0:    
        # Move servo on channel 2 between extremes.
        pwm.set_pwm(2, 0, servo_min)
        time.sleep(1)
        pwm.set_pwm(0, 0, servo_max)
        time.sleep(1)
        print("Servo 2 Completed! ")
        loopCount -= 1

elif sys.argv[1] == "3":
    loopCount = sys.argv[2]
    while loopCount is not 0:
        # Move servo on channel 3 between extremes.
        pwm.set_pwm(3, 0, servo_min)
        time.sleep(1)
        pwm.set_pwm(0, 0, servo_max)
        time.sleep(2)
        print("Servo 3 Completed! ")
        loopCount -= 1

elif sys.argv[1] == "4":
    loopCount = sys.argv[2]
    while loopCount is not 0:
        # Move servo on channel 4 between extremes.
        pwm.set_pwm(4, 0, servo_min)
        time.sleep(1)
        pwm.set_pwm(0, 0, servo_max)
        time.sleep(2)
        print("Servo 4 Completed! ")
        loopCount -= 1
    
elif sys.argv[1] == "5":
    loopCount = sys.argv[2]
    while loopCount is not 0:
        # Move servo on channel 5 between extremes.
        pwm.set_pwm(5, 0, servo_min)
        time.sleep(1)
        pwm.set_pwm(0, 0, servo_max)
        time.sleep(2)
        loopCount -= 1
        print("Servo 5 Completed! ")

efif sys.argv[1] == "mix":
    # Turn mixing motor to mix drink
    pwm.set_pwm(6, 0, servo_min)
    time.sleep(1)
    pwm.set_pwm(0, 0, servo_max)
    time.sleep(2)
    print("Mixing Completed! ")
    