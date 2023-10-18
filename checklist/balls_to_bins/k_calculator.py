import math
import numpy as np

def main():

	k = 1000
	n = 2000
	B = 8

	# answer what value of k makes p = 2^(-40)
	total_sum = 0.
	for i in range(0, k):

		term1 = np.float128(math.comb(n, i))

		term2 = 1./B
		term2 = term2**i

		term3 = 1. - (1./B)
		term3 = term3**(n - i)

		total_sum += (term1*term2*term3)

		print(total_sum)

	total_sum = total_sum**B

	p = 1 - total_sum

	print(p)
	print(2**(-40))



if __name__ == '__main__':
	main()