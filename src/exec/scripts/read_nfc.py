import RPi.GPIO as GPIO
import MFRC522
import sys

from .smbus2.smbus2 import SMBus, i2c_msg

nfcMode = sys.argv[1]
MIFAREReader = MFRC522.MFRC522()
readData= -1
bus = SMBus(1)

try:
    #Read NFC UID for identification
    if nfcMode == 0:
        while readData == -1:
            # Scan for cards
            (status,TagType) = MIFAREReader.MFRC522_Request(MIFAREReader.PICC_REQIDL)

            # If a card is found
            if status == MIFAREReader.MI_OK:
                # Get the UID of the card
                (status,uid) = MIFAREReader.MFRC522_Anticoll()
                # This is the default key for authentication
                key = [0xFF,0xFF,0xFF,0xFF,0xFF,0xFF]
                # Select the scanned tag
                MIFAREReader.MFRC522_SelectTag(uid)
                # Authenticate
                status = MIFAREReader.MFRC522_Auth(MIFAREReader.PICC_AUTHENT1A, 8, key, uid)
                # Check if authenticated
                if status == MIFAREReader.MI_OK:
                    # Read block 8
                    readData = MIFAREReader.MFRC522_Read(8)

    #Read RFID for payment
    elif nfcMode == 1:
        while readData == -1
            # Scan for cards
            time.sleep(0.5)

            readBuffer = []
            read = i2c_msg.read(0x24, 64) #read length 64 from pn532 address
            bus.i2c_rdwr(read)

            for byte in read:
                readBuffer.append(byte)

            readData = readBuffer

    return readData

except KeyboardInterrupt:
    print("Abbruch")
    GPIO.cleanup()