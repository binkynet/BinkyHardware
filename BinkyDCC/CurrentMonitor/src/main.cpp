/*
 * Created on Sat Mar 16 2019
 *
 * Copyright (c) 2019 Ewout Prangsma
 */

#include <Arduino.h>
#include <Wire.h>
#include <LiquidCrystal_I2C.h>

void turnPowerOn();
void turnPowerOff();
void showPercentage();

#define VERSION "0.1"

#define SCL_PHYS_PIN 7
#define SDA_PHYS_PIN 5
#define EN_PORT 3          // digital port to turn booster on/off
#define ENREQ_PORT 5       // digital port to read if DCC central wants the booster turned on/off
#define MAXC_CHAN A1       // ADC channel connected to pot-meter (pin A1)
#define CURB_CHAN A0       // ADC channel connected to current output of H-Bridge (pin A0)
float potReading = 0;
int currentReading = 0;
int enableReq = 0;
bool waitForNotEnabledReq = true;

LiquidCrystal_I2C lcd(0x27, 16, 2); // set address & 16 chars / 2 lines
unsigned long now;
long cAverage = 0;
int avgTimes = 5;
int lastAverage = 0;
float percentage = 0;

void setup()
{
  pinMode(EN_PORT, OUTPUT);
  pinMode(ENREQ_PORT, INPUT);

  lcd.init();        // initialize the lcd
  lcd.backlight();
  lcd.clear(); // Print a message to the LCD.

  lcd.setCursor(0, 0);
  lcd.print("BinkyDCC Booster");
  lcd.setCursor(0, 1);
  lcd.print("Version ");
  lcd.print(VERSION);
  delay(1000);
  lcd.clear();
  turnPowerOn();
  delay(500);
  now = millis();
}

void loop()
{
  potReading = analogRead(MAXC_CHAN);
  potReading = potReading / 100;

  lcd.setCursor(12, 0);
  lcd.print(potReading, 1);
  lcd.print("   ");
  lcd.setCursor(7, 0);
  lcd.print("MaxA=");
  showPercentage();

  cAverage = 0;
  for (int xx = 0; xx < avgTimes; xx++)
  {
    enableReq = digitalRead(ENREQ_PORT);
    if (!enableReq) {
      waitForNotEnabledReq = false;
      turnPowerOff();
    }
    currentReading = analogRead(CURB_CHAN);
    if (currentReading >= 1000)
    {
      // Current is to high, turn off
      turnPowerOff();
      waitForNotEnabledReq = true;
    }
    // Calculate average
    cAverage = cAverage + currentReading;
  }
  currentReading = cAverage / avgTimes;
  if (currentReading != lastAverage)
  {
    if (millis() - now >= 450)
    {
      // only update LCD every 1/2 second to limit flicker
      lcd.setCursor(0, 0);
      lcd.print("C=");
      lcd.print(currentReading);
      lcd.print("  ");
    }
  }
  if ((enableReq) && (!waitForNotEnabledReq)) {
    turnPowerOn();
  }
  lastAverage = currentReading; // keep for compare & print
}

void showPercentage()
{
  if (potReading == 0.0) {
    percentage = 0;
  } else {
    percentage = (currentReading * 0.0105) / potReading; // was 0.014
    percentage = percentage * 100;
  }
  if (millis() - now >= 500) // only update LCD every 1/2 second to limit flicker
  {
    lcd.setCursor(9, 1);
    lcd.print(percentage, 1);
    lcd.print("%  ");
    now = millis();
  }
  if (percentage > 100)
  {
    waitForNotEnabledReq = true;
    turnPowerOff();
  }
}

void turnPowerOff()
{
  digitalWrite(EN_PORT, LOW);
  lcd.setCursor(0, 1);
  if (waitForNotEnabledReq) {
    lcd.print("PWR OFF");
  } else {
    lcd.print("PWR Off");
  }
/*  delay(2000);
  turnPowerOn();
  lcd.setCursor(0, 1);
  lcd.print("               ");*/
}

void turnPowerOn()
{
  digitalWrite(EN_PORT, HIGH);
  lcd.setCursor(0, 1);
  lcd.print("PWR On ");
}
