def calc(base_price)
  yearly_percent = 3
  down_payment_percent = 20
  down_payment = (base_price * (down_payment_percent / 100.0))
  years_to_pay = 7

  final_price =
    base_price + (base_price * (yearly_percent / 100.0) * years_to_pay)

  monthly_payment = (final_price - down_payment) / (years_to_pay * 12)

  [down_payment.round(2), monthly_payment.round(2), final_price.round(2)]
end

def compare(prices)
  prices.map do |price|
    res = calc(price)
    {
      base_price: price,
      down_payment: res[0],
      monthly_payment: res[1],
      final_price: res[2]
    }
  end
end

ap compare([729_000, 1_099_000, 1_362_000, 1_419_000])

[
  30_012,
  49_973,
  49_995,
  43_000,
  43_000,
  40_000,
  20_001,
  10_000
].reduce(:+)

(1..100).map do |i|
  res = i * 10
  next if res > 50
  res
end.compact
